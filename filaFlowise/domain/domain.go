package domain

import "time"

// Request estrutura usada para requisição do Flowise
type Request struct {
	Body       map[string]interface{} `json:"body"`
	UserNs     string                 `json:"userNs"`
	URLFlowise string                 `json:"urlFlowise"`
}

// Response estrutura de resposta padrão
type Response struct {
	Data  map[string]interface{}
	Error error
}

// QueueItem item a ser enfileirado no Redis
type QueueItem struct {
	Request   Request
	Timestamp time.Time
}

// Queue interface para manipulação da fila
type Queue interface {
	Enqueue(userNs string, item QueueItem) error
	Dequeue(userNs string) (QueueItem, error)
	TryLock(userNs string, ttl time.Duration) (bool, error)
	ReleaseLock(userNs string) error
	IsLocked(userNs string) (bool, error)
}
