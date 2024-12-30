////////////////////////////////////////////////////////////////////////////////
// main.go
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/RMS-SH/acumuladora/controllers"
	"github.com/RMS-SH/acumuladora/infrastructure"
	"github.com/RMS-SH/acumuladora/repositories"
	"github.com/RMS-SH/acumuladora/usecases"
	"github.com/joho/godotenv"
)

// main é o ponto de entrada da aplicação.
// Ele carrega variáveis de ambiente, inicializa o MongoDB, configura
// os repositórios, os casos de uso e os controladores, e inicia o servidor HTTP.
func main() {
	// Carregar variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum arquivo .env encontrado ou erro ao carregá-lo")
	} else {
		log.Println("Arquivo .env carregado com sucesso")
	}

	ctx := context.Background()

	// Inicializar cliente MongoDB
	mongoClient := infrastructure.NewMongoDBClient(ctx)

	// Nome do banco de dados (pode ser parametrizável via .env, se desejado)
	dbName := "acumulador"

	// Configurar repositórios
	mongoRepo := repositories.NewMongoDBRepository(mongoClient, dbName, ctx)

	// Configurar casos de uso
	requestUsecase := usecases.NewRequestUsecase(mongoRepo)

	// Ler IPs permitidos da variável de ambiente
	allowedIPsEnv := os.Getenv("ALLOWED_IPS")
	var allowedIPs []string
	if allowedIPsEnv != "" {
		allowedIPs = strings.Split(allowedIPsEnv, ",")
	} else {
		log.Println("Nenhum IP permitido especificado em ALLOWED_IPS")
	}

	// Configurar controladores
	requestController := controllers.NewRequestController(requestUsecase, allowedIPs)

	// Iniciar servidor HTTP
	infrastructure.StartHTTPServer(requestController)
}
