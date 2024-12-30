////////////////////////////////////////////////////////////////////////////////
// controllers/request_controller.go
////////////////////////////////////////////////////////////////////////////////

package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/RMS-SH/acumuladora/usecases"
)

// RequestController é responsável por lidar com as requisições HTTP,
// delegando a lógica de negócio para o RequestUsecase.
type RequestController struct {
	RequestUsecase *usecases.RequestUsecase
	AllowedIPs     []string
}

// NewRequestController cria uma nova instância de RequestController.
func NewRequestController(usecase *usecases.RequestUsecase, allowedIPs []string) *RequestController {
	return &RequestController{
		RequestUsecase: usecase,
		AllowedIPs:     allowedIPs,
	}
}

// HandleRequest processa requisições HTTP recebidas para acumular ou enviar imediatamente os dados.
func (c *RequestController) HandleRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida em /request/")

	// Limitar o tamanho do corpo da requisição a 10KB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	// Extrair access_token do header da requisição
	accessToken := r.Header.Get("access_token")
	if accessToken == "" {
		http.Error(w, "Access token está ausente", http.StatusUnauthorized)
		return
	}

	// Processar a requisição
	statusCode, err := c.RequestUsecase.ProcessRequest(accessToken, r)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleUpdateMinutos atualiza o campo 'minutos' em um registro específico.
func (c *RequestController) HandleUpdateMinutos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Limitar o tamanho do corpo da requisição
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	var updateData struct {
		NomeWorkspace string  `json:"nomeWorkspace"`
		Data          string  `json:"data"`
		Minutos       float64 `json:"minutos"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateData); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if updateData.NomeWorkspace == "" || updateData.Data == "" {
		http.Error(w, "nomeWorkspace e data são obrigatórios", http.StatusBadRequest)
		return
	}

	if err := c.RequestUsecase.UpdateMinutos(updateData.NomeWorkspace, updateData.Data, updateData.Minutos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleAddResponse incrementa o contador de respostas para um determinado workspace/data.
func (c *RequestController) HandleAddResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	var responseData struct {
		NomeWorkspace string `json:"nomeWorkspace"`
		Data          string `json:"data"`
		Count         int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&responseData); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := c.RequestUsecase.AddResponse(responseData.NomeWorkspace, responseData.Data, responseData.Count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleCountImage incrementa o contador de imagens para um determinado workspace/data.
func (c *RequestController) HandleCountImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024) // 10KB

	var countData struct {
		NomeWorkspace string `json:"nomeWorkspace"`
		Data          string `json:"data"`
		Count         int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&countData); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if err := c.RequestUsecase.CountImage(countData.NomeWorkspace, countData.Data, countData.Count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
