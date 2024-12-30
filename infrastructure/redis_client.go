////////////////////////////////////////////////////////////////////////////////
// infrastructure/redis_client.go
////////////////////////////////////////////////////////////////////////////////

package infrastructure

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient cria e retorna um cliente Redis a partir das variáveis de ambiente.
func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // Valor padrão, se não setado
	}

	password := os.Getenv("REDIS_PASSWORD") // Pode ser vazio
	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		dbStr = "0"
	}
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Printf("Valor inválido para REDIS_DB, usando DB=0: %v", err)
		db = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	log.Printf("Conectado ao Redis em %s, DB=%d", addr, db)
	return client
}
