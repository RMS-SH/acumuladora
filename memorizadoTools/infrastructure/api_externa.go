// infrastructure/api_externa.go
package infrastructure

import (
	"github.com/RMS-SH/acumuladora/memorizadoTools/config"
	"log"
	"net/http"
	"time"
)

// NovoAPIExternaCliente configura um http.Client com timeout para chamadas externas
func NovoAPIExternaCliente(cfg *config.Config) *http.Client {
	log.Printf("Criando cliente HTTP para API Externa com timeout de %d segundos.", cfg.APIExternaTimeout)
	return &http.Client{
		Timeout: time.Duration(cfg.APIExternaTimeout) * time.Second,
	}
}
