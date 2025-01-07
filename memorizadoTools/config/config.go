// config/config.go
package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config representa as configurações da aplicação
type Config struct {
	RedisURL          string
	APIExternaURL     string
	APIExternaTimeout int
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

// LoadConfig carrega as variáveis de ambiente do .env ou do sistema
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado ou não carregado, usando variáveis de ambiente do sistema.")
	} else {
		log.Println("Variáveis de ambiente do arquivo .env carregadas.")
	}

	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	log.Printf("RedisURL: %s", redisURL)

	apiExternaURL := getEnv("API_EXTERNA_URL", "https://r001.api.datarms.com/webhook/KSOL_consultaImoveis")
	log.Printf("APIExternaURL (fallback): %s", apiExternaURL)

	// Timeout para chamadas externas
	timeout := 360
	if val, err := strconv.Atoi(getEnv("API_EXTERNA_TIMEOUT", "30")); err == nil && val > 0 {
		timeout = val
	}
	log.Printf("APIExternaTimeout: %d segundos", timeout)

	return &Config{
		RedisURL:          redisURL,
		APIExternaURL:     apiExternaURL,
		APIExternaTimeout: timeout,
	}
}
