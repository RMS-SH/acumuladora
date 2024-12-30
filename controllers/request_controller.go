////////////////////////////////////////////////////////////////////////////////
// controllers/request_controller.go
////////////////////////////////////////////////////////////////////////////////

package controllers

import (
	"log"
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
	log.Println("Requisição recebida em /request/")

	// Limitar o tamanho do corpo da requisição a 10KB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	// Processar a requisição
	statusCode, err := c.RequestUsecase.ProcessRequest(r)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}
