// ==================================================
// infrastructure/flowise/client.go
// ==================================================
package flowise

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Client interface para o cliente Flowise
type Client interface {
	SendRequest(url string, body map[string]interface{}) (map[string]interface{}, error)
}

// FlowiseClient implementação de Client
type FlowiseClient struct {
	httpClient *http.Client
}

// NewFlowiseClient cria uma nova instância do FlowiseClient
func NewFlowiseClient() *FlowiseClient {
	log.Println("Inicializando FlowiseClient com cliente HTTP padrão.")
	return &FlowiseClient{
		httpClient: &http.Client{},
	}
}

// SendRequest envia uma requisição POST para a URL especificada com o corpo fornecido
func (c *FlowiseClient) SendRequest(url string, body map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Enviando requisição para Flowise em %s com corpo: %v", url, body)
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Printf("Erro ao serializar o corpo da requisição: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Erro ao criar a requisição HTTP: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar a requisição para Flowise: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Recebida resposta da Flowise com status: %s", resp.Status)

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		log.Printf("Erro ao decodificar a resposta da Flowise: %v", err)
		return nil, err
	}

	log.Printf("Resposta da Flowise: %v", responseBody)
	return responseBody, nil
}
