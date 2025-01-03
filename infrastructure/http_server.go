////////////////////////////////////////////////////////////////////////////////
// infrastructure/http_server.go
////////////////////////////////////////////////////////////////////////////////

package infrastructure

import (
	"log"
	"net/http"
	"time"

	"github.com/RMS-SH/acumuladora/controllers"
)

// StartHTTPServer inicia o servidor HTTP configurando apenas a rota /request/.
func StartHTTPServer(controller *controllers.RequestController) {
	server := &http.Server{
		Addr:              ":7511",
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		ReadHeaderTimeout: 5 * time.Second,
	}

	http.HandleFunc("/request/", controller.HandleRequest)

	log.Println("Servidor HTTP rodando na porta 7511")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Falha no servidor HTTP: %v", err)
	}
}
