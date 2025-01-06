////////////////////////////////////////////////////////////////////////////////
// controllers/request_controller.go
////////////////////////////////////////////////////////////////////////////////

package controllers

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/RMS-SH/acumuladora/usecases"
)

// RequestController é responsável por lidar com as requisições HTTP da rota /request/.
type RequestController struct {
	RequestUsecase *usecases.RequestUsecase
}

// NewRequestController cria uma nova instância de RequestController.
func NewRequestController(usecase *usecases.RequestUsecase) *RequestController {
	return &RequestController{
		RequestUsecase: usecase,
	}
}

// HandleRequest processa requisições HTTP recebidas para acumular ou enviar imediatamente os dados.
func (c *RequestController) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Limitar o tamanho do corpo da requisição a 10KB (conforme pedido)
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	// ========================= LOGS DE ENTRADA =========================
	// 1. Log de quando o request é recebido
	// 1.1 Tamanho e IP que mandou o request

	// Tentar pegar IP do cabeçalho "X-Forwarded-For" (recomendado quando atrás de proxy),
	// se vazio, usar r.RemoteAddr
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	// O RemoteAddr pode conter a porta junto, ex.: "192.168.0.10:54321".
	// Então podemos extrair somente o IP se quisermos.
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	// ContentLength retorna o tamanho do corpo que o cliente informou no header,
	// caso não seja -1. Se for -1, não há informação.
	contentLength := r.ContentLength
	if contentLength < 0 {
		contentLength = 0
	}

	log.Printf("[ENTRADA] Requisição recebida em /request/. IP: %s, Tamanho: %d bytes", ip, contentLength)
	// ========================= FIM LOGS DE ENTRADA =========================

	// Processar a requisição
	statusCode, err := c.RequestUsecase.ProcessRequest(r)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Request processado com sucesso")
}
