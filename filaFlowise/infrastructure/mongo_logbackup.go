// ==================================================
// infrastructure/mongo_logbackup.go
// ==================================================
package infrastructure

import (
	"context"
	"log"
	"time"

	"github.com/RMS-SH/filaFlowise/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoLogBackup implementação de LogBackup utilizando MongoDB
type MongoLogBackup struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoLogBackup cria uma nova instância de MongoLogBackup
func NewMongoLogBackup(mongoURI, dbName, collectionName string) (*MongoLogBackup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)

	return &MongoLogBackup{
		client:     client,
		collection: collection,
	}, nil
}

// SaveFailedRequest salva logs de requisições com falha no MongoDB
func (m *MongoLogBackup) SaveFailedRequest(logData domain.FailedRequestLog) error {
	logData.Timestamp = time.Now()
	_, err := m.collection.InsertOne(context.Background(), bson.M{
		"userNs":       logData.UserNs,
		"request":      logData.Request,
		"responseData": logData.ResponseData,
		"errorMsg":     logData.ErrorMsg,
		"timestamp":    logData.Timestamp,
	})
	if err != nil {
		log.Printf("Erro ao salvar log de falha no MongoDB: %v", err)
	}
	return err
}
