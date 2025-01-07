// domain/domain.go
package domain

// Requisicao contém as informações enviadas no body da rota /requisicao
type Requisicao struct {
	UserNs            string                 `json:"userNs"`
	APIExternaURL     string                 `json:"apiExternaURL"`
	Dados             map[string]interface{} `json:"dados"`
	ExpiracaoSegundos int                    `json:"expiracaoSegundos"` // TTL dinâmico (se > 0)
}
