// ==================================================
// infrastructure/redis.go
// ==================================================
package infrastructure

import (
	"context"
	"github.com/RMS-SH/acumuladora/filaFlowise/config"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

// NewRedisClient cria e retorna um novo cliente Redis
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	log.Printf("Configurando cliente Redis com URL: %s", cfg.RedisURL)
	options, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("Erro ao analisar a URL do Redis: %v", err)
		return nil, err
	}

	client := redis.NewClient(options)

	// Verifica a conexão com o Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Verificando conexão com Redis...")
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Erro ao conectar ao Redis: %v", err)
		return nil, err
	}

	log.Println("Conexão com Redis verificada com sucesso.")
	return client, nil
}
