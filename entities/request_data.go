////////////////////////////////////////////////////////////////////////////////
// entities/request_data.go
////////////////////////////////////////////////////////////////////////////////

package entities

// RequestData representa a estrutura de cada requisição recebida.
// Pode conter cabeçalhos, parâmetros, query strings, corpo em forma de slice
// de BodyItem, bem como um webhook de callback e modo de execução.
type RequestData struct {
	Headers       map[string]string `json:"headers"`
	Params        map[string]string `json:"params"`
	Query         map[string]string `json:"query"`
	Body          []BodyItem        `json:"body"`
	WebhookUrl    string            `json:"webhookUrl"`
	ExecutionMode string            `json:"executionMode"`
}
