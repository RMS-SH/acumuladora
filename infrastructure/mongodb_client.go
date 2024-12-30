////////////////////////////////////////////////////////////////////////////////
// infrastructure/mongodb_client.go
////////////////////////////////////////////////////////////////////////////////

package infrastructure

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoDBClient inicializa e retorna um novo cliente MongoDB.
// Ele se conecta ao servidor MongoDB definido pela variável de ambiente MONGODB_URI.
func NewMongoDBClient(ctx context.Context) *mongo.Client {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("A variável de ambiente MONGODB_URI não está definida")
	}

	clientOptions := options.Client().ApplyURI(mongoURI).SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Falha ao conectar ao MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Falha ao realizar ping no MongoDB: %v", err)
	}

	log.Println("Conexão estabelecida com sucesso ao MongoDB")
	return client
}
