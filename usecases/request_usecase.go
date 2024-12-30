////////////////////////////////////////////////////////////////////////////////
// usecases/request_usecase.go
////////////////////////////////////////////////////////////////////////////////

package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RMS-SH/acumuladora/entities"
	"github.com/RMS-SH/acumuladora/repositories"
)

// RequestUsecase concentra a lógica de negócio relacionada ao processamento
// das requisições e ao envio (ou acumulação) dos dados.
type RequestUsecase struct {
	MongoRepo *repositories.MongoDBRepository

	timers   map[string]*time.Timer
	timersMu sync.Mutex

	accessTokenCache   map[string]time.Time
	accessTokenCacheMu sync.Mutex
}

// NewRequestUsecase cria uma instância de RequestUsecase com as estruturas internas
// para o agendamento de timers (acúmulo) e cache de tokens de acesso.
func NewRequestUsecase(mongoRepo *repositories.MongoDBRepository) *RequestUsecase {
	return &RequestUsecase{
		MongoRepo:          mongoRepo,
		timers:             make(map[string]*time.Timer),
		timersMu:           sync.Mutex{},
		accessTokenCache:   make(map[string]time.Time),
		accessTokenCacheMu: sync.Mutex{},
	}
}

// ProcessRequest lida com a análise do token de acesso, identificação do tempo
// de acúmulo (caso exista) e despacho das requisições para envio imediato ou
// salvamento para envio posterior.
func (u *RequestUsecase) ProcessRequest(accessToken string, r *http.Request) (int, error) {
	// 1. Validar token de acesso.
	valid, err := u.ValidateAccessToken(accessToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !valid {
		return http.StatusUnauthorized, errors.New("Access token inválido")
	}

	// 2. Extrair parâmetro de tempo (seconds) da URL (/request/{timeParam}).
	timeParam, err := u.parseTimeParam(r.URL.Path)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// 3. Ler e interpretar o corpo da requisição como um ou múltiplos RequestData.
	requestDataList, err := u.parseRequestBody(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// 4. Processar cada RequestData recebido.
	for _, requestData := range requestDataList {
		userNS, url, err := extractUserNSData(requestData)
		if err != nil {
			return http.StatusBadRequest, err
		}

		// 4.1 Incrementar contagem de requests recebidos no workspace correspondente.
		nomeWorkspace := requestData.Body[0].NomeWorkspace
		currentDate := time.Now().Format("2006-01-02")
		if err := u.MongoRepo.IncrementCounters(nomeWorkspace, currentDate, map[string]int{"requestRecebidos": 1}); err != nil {
			log.Printf("Erro ao incrementar contadores: %v", err)
			return http.StatusInternalServerError, err
		}

		// 4.2 Se timeParam == 0, envia os dados imediatamente; caso contrário, acumula.
		if timeParam == 0 {
			// Envio imediato dos dados (garantindo que o item 'dados' seja enviado primeiro).
			dataToSend := reorderDataToSend(requestData.Body)
			if err := u.sendDataToURL(url, dataToSend); err != nil {
				log.Printf("Erro ao enviar dados para a URL: %v", err)
				return http.StatusInternalServerError, err
			}
		} else {
			// Salvar dados no MongoDB para envio posterior.
			if err := u.MongoRepo.SaveUserData(userNS, requestData.Body, url); err != nil {
				log.Printf("Erro ao salvar dados do usuário: %v", err)
				return http.StatusInternalServerError, err
			}
			// Agendar processamento dos dados após o tempo de acúmulo.
			u.scheduleDataProcessing(userNS, timeParam+1)
		}
	}

	return http.StatusOK, nil
}

// ValidateAccessToken verifica se o token de acesso é válido. Primeiro consulta
// o cache local, caso não encontre ou esteja expirado, faz a verificação no MongoDB.
func (u *RequestUsecase) ValidateAccessToken(accessToken string) (bool, error) {
	u.accessTokenCacheMu.Lock()
	expiration, found := u.accessTokenCache[accessToken]
	u.accessTokenCacheMu.Unlock()

	if found && time.Now().Before(expiration) {
		return true, nil
	}

	isValid, err := u.MongoRepo.IsAccessTokenValid(accessToken)
	if err != nil {
		return false, err
	}
	if isValid {
		u.accessTokenCacheMu.Lock()
		u.accessTokenCache[accessToken] = time.Now().Add(60 * time.Minute)
		u.accessTokenCacheMu.Unlock()
	}
	return isValid, nil
}

// scheduleDataProcessing agenda um timer para envio dos dados acumulados
// de um determinado userNS após 'delaySeconds' segundos.
func (u *RequestUsecase) scheduleDataProcessing(userNS string, delaySeconds int) {
	u.timersMu.Lock()
	defer u.timersMu.Unlock()

	// Se já existe um timer para o userNS, cancelá-lo e criar um novo.
	if existingTimer, exists := u.timers[userNS]; exists {
		existingTimer.Stop()
	}

	duration := time.Duration(delaySeconds) * time.Second
	timer := time.AfterFunc(duration, func() {
		u.processUserData(userNS)
	})

	u.timers[userNS] = timer
}

// processUserData recupera os dados acumulados no MongoDB para um userNS,
// envia esses dados à URL armazenada e remove o registro do banco.
func (u *RequestUsecase) processUserData(userNS string) {
	u.timersMu.Lock()
	delete(u.timers, userNS)
	u.timersMu.Unlock()

	userData, err := u.MongoRepo.FetchUserData(userNS)
	if err != nil {
		log.Printf("Erro ao buscar dados do usuário para userNS %s: %v", userNS, err)
		return
	}

	dataToSend := reorderDataToSend(userData.Body)
	if err := u.sendDataToURL(userData.URL, dataToSend); err != nil {
		log.Printf("Erro ao enviar dados para a URL para userNS %s: %v", userNS, err)
	}

	if err := u.MongoRepo.DeleteUserData(userNS); err != nil {
		log.Printf("Erro ao deletar dados do usuário para userNS %s: %v", userNS, err)
	}
}

// sendDataToURL faz a serialização para JSON e o envio HTTP (POST) dos dados.
func (u *RequestUsecase) sendDataToURL(url string, dataToSend []entities.BodyItem) error {
	jsonData, err := json.Marshal(dataToSend)
	if err != nil {
		return fmt.Errorf("Erro ao codificar JSON: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("Erro ao criar requisição HTTP: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Erro ao enviar requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("código de status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Dados enviados com sucesso para a URL: %s", url)
	return nil
}

// parseTimeParam tenta extrair o timeParam do caminho da URL. Exemplo:
// /request/10 extrai 10 como timeParam (segundos).
func (u *RequestUsecase) parseTimeParam(path string) (int, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, errors.New("Tempo de espera não especificado na URL")
	}
	timeParam, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, errors.New("Tempo de espera inválido")
	}
	return timeParam, nil
}

// parseRequestBody tenta desserializar o corpo da requisição em um slice
// de RequestData, ou em um único RequestData, ou ainda em um slice de BodyItem.
func (u *RequestUsecase) parseRequestBody(r *http.Request) ([]entities.RequestData, error) {
	var requestDataList []entities.RequestData

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("Erro ao ler o corpo da requisição")
	}

	// 1. Tentar desserializar como []RequestData
	if err := json.Unmarshal(bodyBytes, &requestDataList); err == nil {
		return requestDataList, nil
	}

	// 2. Tentar desserializar como RequestData
	var singleRequestData entities.RequestData
	if err := json.Unmarshal(bodyBytes, &singleRequestData); err == nil && len(singleRequestData.Body) > 0 {
		return []entities.RequestData{singleRequestData}, nil
	}

	// 3. Tentar desserializar como []BodyItem
	var bodyItems []entities.BodyItem
	if err := json.Unmarshal(bodyBytes, &bodyItems); err == nil && len(bodyItems) > 0 {
		singleRequestData.Body = bodyItems
		return []entities.RequestData{singleRequestData}, nil
	}

	return nil, errors.New("Corpo da requisição inválido ou vazio")
}

// AddResponse incrementa o contador de "requestEncaminhados" em um workspace/data.
func (u *RequestUsecase) AddResponse(nomeWorkspace, date string, count int) error {
	return u.MongoRepo.IncrementResponseCounter(nomeWorkspace, date, count)
}

// CountImage incrementa o contador de imagens em um workspace/data.
func (u *RequestUsecase) CountImage(nomeWorkspace, date string, count int) error {
	return u.MongoRepo.IncrementImageCounter(nomeWorkspace, date, count)
}

// UpdateMinutos atualiza o campo minutos em um workspace/data.
func (u *RequestUsecase) UpdateMinutos(nomeWorkspace, date string, minutos float64) error {
	return u.MongoRepo.UpdateMinutos(nomeWorkspace, date, minutos)
}

////////////////////////////////////////////////////////////////////////////////
// Funções auxiliares
////////////////////////////////////////////////////////////////////////////////

// extractUserNSData extrai o userNS e URL do item "dados".
// Retorna erro caso os campos estejam ausentes.
func extractUserNSData(rd entities.RequestData) (userNS string, url string, err error) {
	for _, item := range rd.Body {
		if item.Type == "dados" {
			userNS = item.UserNS
			url = item.URL
			break
		}
	}
	if userNS == "" || url == "" {
		return "", "", errors.New("userNS ou URL ausente no item 'dados'")
	}
	return userNS, url, nil
}

// reorderDataToSend garante que o item com Type="dados" seja enviado primeiro.
func reorderDataToSend(bodyItems []entities.BodyItem) []entities.BodyItem {
	var dadosItem entities.BodyItem
	var outrosItens []entities.BodyItem

	for _, item := range bodyItems {
		if item.Type == "dados" {
			dadosItem = item
		} else {
			outrosItens = append(outrosItens, item)
		}
	}
	if dadosItem.Type == "" {
		// Se não houver um item com Type="dados", apenas retorna o array original
		return bodyItems
	}
	return append([]entities.BodyItem{dadosItem}, outrosItens...)
}