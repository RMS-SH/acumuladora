// interfaces/handler_requisicao.go
package interfaces

import (
	"encoding/json"
	"github.com/RMS-SH/acumuladora/memorizadoTools/domain"
	"github.com/RMS-SH/acumuladora/memorizadoTools/usecase"
	"io"
	"log"
	"net/http"
)

// RequisicaoHandler manipula requisições para a rota /requisicao
type RequisicaoHandler struct {
	useCase usecase.RequisicaoUseCase
}

// NovoRequisicaoHandler cria uma nova instância de RequisicaoHandler
func NovoRequisicaoHandler(u usecase.RequisicaoUseCase) *RequisicaoHandler {
	return &RequisicaoHandler{
		useCase: u,
	}
}

// ProcessarRequisicao lida com POST /requisicao
func (h *RequisicaoHandler) ProcessarRequisicao(w http.ResponseWriter, r *http.Request) {
	log.Println("Recebida requisição no endpoint /requisicao")

	// Ler corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Erro ao ler body da requisição: %v", err)
		http.Error(w, "Body inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Deserializar JSON para domain.Requisicao
	var req domain.Requisicao
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Erro ao parsear JSON: %v", err)
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Valida campos obrigatórios
	if req.UserNs == "" {
		log.Println("Campo 'userNs' é obrigatório")
		http.Error(w, "Campo 'userNs' é obrigatório", http.StatusBadRequest)
		return
	}
	if req.APIExternaURL == "" {
		log.Println("Campo 'apiExternaURL' é obrigatório")
		http.Error(w, "Campo 'apiExternaURL' é obrigatório", http.StatusBadRequest)
		return
	}

	// Chama o UseCase para processar a requisição
	log.Printf("Processando requisicao para userNs=%s, expiracaoSegundos=%d, urlExterna=%s",
		req.UserNs, req.ExpiracaoSegundos, req.APIExternaURL)

	respBytes, err := h.useCase.ProcessarRequisicao(r.Context(), &req)
	if err != nil {
		log.Printf("Erro ao processar requisição: %v", err)
		http.Error(w, "Erro interno ao processar requisição", http.StatusInternalServerError)
		return
	}

	// Retorna resposta da API externa (ou do repositório) para o cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
