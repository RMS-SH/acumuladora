// ==================================================
// usecase/processor.go
// ==================================================
package usecase

import (
	"context"
	"errors"
	"github.com/RMS-SH/acumuladora/filaFlowise/domain"
	"github.com/RMS-SH/acumuladora/filaFlowise/infrastructure/flowise"
	"log"
	"time"
)

// Processor gerencia o fluxo de processamento, impedindo concorrência para o mesmo userNs
type Processor struct {
	Queue     domain.Queue
	Client    flowise.Client
	LogBackup domain.LogBackup
	Timeout   time.Duration
	LockTTL   time.Duration
}

// NewProcessor cria uma nova instância de Processor
func NewProcessor(queue domain.Queue, client flowise.Client, logBackup domain.LogBackup, timeout time.Duration) *Processor {
	log.Println("Inicializando Processor para a fila.")
	return &Processor{
		Queue:     queue,
		Client:    client,
		LogBackup: logBackup,
		Timeout:   timeout,
		LockTTL:   5 * time.Minute, // Exemplo de lock TTL
	}
}

// ProcessRequest controla o fluxo de aquisição de lock, processamento e liberação de lock
func (p *Processor) ProcessRequest(request domain.Request) (*domain.Response, error) {
	log.Printf("Iniciando processamento para userNs=%s.", request.UserNs)
	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout)
	defer cancel()

	// Tentamos adquirir o lock a cada 10 segundos até estourar o timeout
	for {
		select {
		case <-ctx.Done():
			log.Printf("Timeout ao aguardar disponibilidade de lock para userNs=%s.", request.UserNs)
			return nil, errors.New("timeout ao aguardar disponibilidade para processamento")
		default:
			locked, err := p.Queue.TryLock(request.UserNs, p.LockTTL)
			if err != nil {
				log.Printf("Erro ao tentar adquirir lock para userNs=%s: %v", request.UserNs, err)
				return nil, err
			}
			if locked {
				log.Printf("Lock adquirido para userNs=%s. Iniciando processamento.", request.UserNs)
				defer func() {
					if err := p.Queue.ReleaseLock(request.UserNs); err != nil {
						log.Printf("Erro ao liberar lock para userNs=%s: %v", request.UserNs, err)
					} else {
						log.Printf("Lock liberado para userNs=%s.", request.UserNs)
					}
				}()

				resp, err := p.processWithRetry(request)
				if err != nil {
					log.Printf("Erro final ao processar requisição para userNs=%s: %v", request.UserNs, err)
					return nil, err
				}
				log.Printf("Requisição processada com sucesso para userNs=%s.", request.UserNs)
				return resp, nil
			}

			log.Printf("Lock em uso para userNs=%s. Aguardando 10 segundos para tentar novamente.", request.UserNs)
			select {
			case <-ctx.Done():
				log.Printf("Timeout durante a espera de lock para userNs=%s.", request.UserNs)
				return nil, errors.New("timeout ao aguardar disponibilidade para processamento")
			case <-time.After(10 * time.Second):
				// Tenta novamente
			}
		}
	}
}

// processWithRetry faz uma chamada ao process(request) e, em caso de falha, tenta uma segunda vez.
// Em caso de nova falha, envia para LogBackup.
func (p *Processor) processWithRetry(request domain.Request) (*domain.Response, error) {
	resp, err := p.process(request)
	if err != nil {
		log.Printf("Falha ao processar requisição na primeira tentativa para userNs=%s: %v", request.UserNs, err)
		secondResp, secondErr := p.process(request)
		if secondErr != nil {
			log.Printf("Falha na segunda tentativa para userNs=%s: %v", request.UserNs, secondErr)
			p.backupFailure(request, nil, secondErr)
			return nil, secondErr
		}
		return secondResp, nil
	}
	return resp, nil
}

// process realiza o envio da requisição para o Flowise e retorna a resposta
func (p *Processor) process(request domain.Request) (*domain.Response, error) {
	log.Printf("Enviando requisição para Flowise para userNs=%s.", request.UserNs)
	transformedBody := transformBody(request.Body)

	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout)
	defer cancel()

	c := make(chan *domain.Response, 1)
	go func() {
		resp, err := p.Client.SendRequest(request.URLFlowise, transformedBody)
		c <- &domain.Response{Data: resp, Error: err}
	}()

	select {
	case res := <-c:
		if res.Error != nil {
			log.Printf("Erro na resposta do Flowise para userNs=%s: %v", request.UserNs, res.Error)
			return nil, res.Error
		}
		log.Printf("Resposta recebida do Flowise para userNs=%s: %v", request.UserNs, res.Data)
		return res, nil
	case <-ctx.Done():
		log.Printf("Timeout ao aguardar resposta do Flowise para userNs=%s.", request.UserNs)
		p.backupFailure(request, nil, errors.New("timeout ao aguardar resposta do Flowise"))
		return nil, errors.New("timeout ao aguardar resposta do Flowise")
	}
}

// backupFailure envia informações da falha para o MongoDB
func (p *Processor) backupFailure(request domain.Request, response map[string]interface{}, err error) {
	log.Printf("Fazendo backup da falha de processamento para userNs=%s no MongoDB.", request.UserNs)
	logData := domain.FailedRequestLog{
		UserNs:       request.UserNs,
		Request:      request.Body,
		ResponseData: response,
		ErrorMsg:     err.Error(),
	}
	if derr := p.LogBackup.SaveFailedRequest(logData); derr != nil {
		log.Printf("Erro ao salvar log de falha no MongoDB: %v", derr)
	}
}

// Exemplo de possível transformação do corpo
func transformBody(body map[string]interface{}) map[string]interface{} {
	// Ajuste ou mapeamento do payload antes de enviar para o Flowise
	log.Printf("Transformando corpo da requisição: %v", body)
	return body
}
