// usecase/requisicao_usecase.go
package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/RMS-SH/acumuladora/memorizadoTools/domain"
	"github.com/RMS-SH/acumuladora/memorizadoTools/repository"
	"io"
	"log"
	"net/http"
	"time"
)

// RequisicaoUseCase define a interface para o caso de uso
type RequisicaoUseCase interface {
	ProcessarRequisicao(ctx context.Context, req *domain.Requisicao) ([]byte, error)
}

type requisicaoUseCase struct {
	repo              repository.RequisicaoRepository
	httpClient        *http.Client
	apiExternaURLBase string        // fallback
	expiracaoDefault  time.Duration // fallback de TTL
}

// NovoRequisicaoUseCase instancia a implementação do caso de uso
func NovoRequisicaoUseCase(
	repo repository.RequisicaoRepository,
	httpClient *http.Client,
	apiExternaURL string,
	expiracao time.Duration,
) RequisicaoUseCase {
	log.Println("Iniciando RequisicaoUseCase...")
	return &requisicaoUseCase{
		repo:              repo,
		httpClient:        httpClient,
		apiExternaURLBase: apiExternaURL,
		expiracaoDefault:  expiracao,
	}
}

// ProcessarRequisicao faz o merge de dados (se houver) e chama a API externa
func (u *requisicaoUseCase) ProcessarRequisicao(ctx context.Context, req *domain.Requisicao) ([]byte, error) {
	log.Printf("Processando requisição para userNs=%s", req.UserNs)

	// Gerar SHA256 da URL externa
	hashURL := GerarSHA256(req.APIExternaURL)
	// Monta a chave composta
	compositeKey := req.UserNs + "-" + hashURL

	// 1. Buscar dados existentes
	dadosExistentes, ts, err := u.repo.ObterDados(ctx, compositeKey)
	if err != nil {
		log.Printf("Erro ao obter dados existentes: %v", err)
		return nil, err
	}

	now := time.Now()
	var dadosAtualizados map[string]interface{}
	var dadosParaSalvar map[string]interface{}

	// Se não há dados ou já passaram do TTL default, usa somente os novos
	if dadosExistentes == nil || now.Sub(ts) > u.expiracaoDefault {
		dadosAtualizados = req.Dados
		dadosParaSalvar = req.Dados
		log.Printf("Nenhum dado válido encontrado ou TTL default expirado: usando dados do body")
	} else {
		// Caso contrário, mescla
		log.Printf("Dados existentes encontrados, mesclando com novos")
		dadosAtualizados = mesclarDados(dadosExistentes, req.Dados)
		// Armazena somente as novas chaves, se preferir. Ou tudo. Aqui, para simplificar, vamos salvar apenas as chaves novas:
		dadosParaSalvar = req.Dados
	}

	// 2. Definir TTL dinâmico
	var ttl time.Duration
	if req.ExpiracaoSegundos > 0 {
		ttl = time.Duration(req.ExpiracaoSegundos) * time.Second
	} else {
		ttl = u.expiracaoDefault
	}

	// 3. Salvar dados no Redis (agora usando compositeKey)
	if err := u.repo.SalvarDados(ctx, compositeKey, dadosParaSalvar, ttl); err != nil {
		log.Printf("Erro ao salvar dados no Redis: %v", err)
		return nil, err
	}
	log.Printf("Dados salvos no Redis com TTL=%v para userNs=%s", ttl, req.UserNs)

	// 4. Chamar a API externa usando a URL recebida do body
	apiURL := req.APIExternaURL
	log.Printf("Chamando API externa em %s para userNs=%s", apiURL, req.UserNs)

	respBytes, err := u.chamarAPIExterna(ctx, apiURL, dadosAtualizados)
	if err != nil {
		log.Printf("Erro ao chamar API externa: %v", err)
		return nil, err
	}

	log.Printf("Requisição completa para userNs=%s. Retornando resposta da API externa.", req.UserNs)
	return respBytes, nil
}

// mesclarDados faz o merge dos dados existentes com novos campos (override)
func mesclarDados(antigos, novos map[string]interface{}) map[string]interface{} {
	for k, v := range novos {
		antigos[k] = v
	}
	return antigos
}

// chamarAPIExterna envia um POST JSON para a URL informada
func (u *requisicaoUseCase) chamarAPIExterna(ctx context.Context, url string, dados map[string]interface{}) ([]byte, error) {
	bodyJSON, err := json.Marshal(dados)
	if err != nil {
		log.Printf("Erro ao serializar dados para API externa: %v", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		log.Printf("Erro ao criar requisição HTTP: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar requisição para API externa: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API externa retornou status %d", resp.StatusCode)
		return nil, errors.New("API externa retornou status não-OK")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler corpo da resposta da API externa: %v", err)
		return nil, err
	}

	return respBody, nil
}
