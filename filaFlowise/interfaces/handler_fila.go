// ==================================================
// interfaces/handler_fila.go
// ==================================================
package interfaces

import (
	"encoding/json"
	"github.com/RMS-SH/acumuladora/filaFlowise/domain"
	"github.com/RMS-SH/acumuladora/filaFlowise/usecase"
	"log"
	"net/http"
)

// FilaHandler manipula as requisições para o endpoint /process
type FilaHandler struct {
	processor *usecase.Processor
}

// NewFilaHandler cria uma nova instância de FilaHandler
func NewFilaHandler(p *usecase.Processor) *FilaHandler {
	return &FilaHandler{
		processor: p,
	}
}

// ProcessRequest lida com a requisição POST para /process
func (h *FilaHandler) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Recebida requisição no endpoint /process.")
	var req domain.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar a requisição: %v", err)
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.UserNs == "" {
		log.Println("Campo 'userNs' ausente na requisição.")
		http.Error(w, "Campo 'userNs' é obrigatório", http.StatusBadRequest)
		return
	}

	log.Printf("Processando requisição para userNs=%s.", req.UserNs)
	response, err := h.processor.ProcessRequest(req)
	if err != nil {
		log.Printf("Erro ao processar requisição: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Requisição processada com sucesso para userNs=%s.", req.UserNs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.Data)
}
