////////////////////////////////////////////////////////////////////////////////
// main.go
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"github.com/RMS-SH/acumuladora/controllers"
	"github.com/RMS-SH/acumuladora/infrastructure"
	"github.com/RMS-SH/acumuladora/repositories"
	"github.com/RMS-SH/acumuladora/usecases"
)

// main é o ponto de entrada da aplicação.
// Carrega variáveis de ambiente, inicializa o Redis, configura o repositório,
// o caso de uso e o controlador para a rota /request/ e inicia o servidor HTTP.
func main() {
	// Carregar variáveis de ambiente do arquivo .env (opcional)
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum arquivo .env encontrado ou erro ao carregá-lo")
	} else {
		log.Println("Arquivo .env carregado com sucesso")
	}

	// Inicializar cliente Redis
	redisClient := infrastructure.NewRedisClient(context.TODO())

	// Configurar repositório
	redisRepo := repositories.NewRedisRepository(redisClient)

	// Configurar caso de uso
	requestUsecase := usecases.NewRequestUsecase(redisRepo)

	// Configurar controlador
	requestController := controllers.NewRequestController(requestUsecase)

	// Iniciar servidor HTTP (rota /request/)
	infrastructure.StartHTTPServer(requestController)
}
