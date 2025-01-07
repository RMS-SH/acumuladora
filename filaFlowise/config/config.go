package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config representa as configurações da aplicação
type Config struct {
	RedisURL string
	MongoURI string
}

// getEnv lê uma variável de ambiente ou retorna um valor padrão
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		// Se a variável de ambiente aponta para um arquivo de secret, ler o conteúdo
		if strings.HasPrefix(value, "/run/secrets/") {
			data, err := os.ReadFile(value)
			if err != nil {
				log.Printf("Erro ao ler secret %s: %v", key, err)
				return fallback
			}
			return strings.TrimSpace(string(data))
		}
		return value
	}
	return fallback
}

// LoadConfig carrega as variáveis de ambiente e retorna uma instância de Config
func LoadConfig() *Config {
	// Carrega variáveis de ambiente do arquivo .env (se existir)
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado ou não carregado, usando variáveis de ambiente do sistema.")
	} else {
		log.Println("Variáveis de ambiente do arquivo .env carregadas.")
	}

	redisURL := getEnv("REDIS_URL", "redis://redis:6379")
	log.Printf("RedisURL: %s", redisURL)

	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	log.Printf("MongoURI: %s", mongoURI)

	return &Config{
		RedisURL: redisURL,
		MongoURI: mongoURI,
	}
}
