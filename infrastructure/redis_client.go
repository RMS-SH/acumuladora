package infrastructure

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient inicializa e retorna um novo cliente Redis.
// Ele se conecta usando a variável de ambiente REDIS_URI no formato:
// redis://<username>:<password>@<host>:<port>/<db>
func NewRedisClient(ctx context.Context) *redis.Client {
	redisURI := os.Getenv("REDIS_URI")
	if redisURI == "" {
		log.Fatal("A variável de ambiente REDIS_URI não está definida")
	}

	opts, err := redis.ParseURL(redisURI)
	if err != nil {
		log.Fatalf("Erro ao analisar a URL do Redis: %v", err)
	}

	client := redis.NewClient(opts)

	// Testar a conexão com um PING
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Falha ao conectar no Redis: %v", err)
	}

	log.Println("Conexão estabelecida com sucesso ao Redis")
	return client
}
