////////////////////////////////////////////////////////////////////////////////
// repositories/mongodb_repository.go
////////////////////////////////////////////////////////////////////////////////

package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/RMS-SH/acumuladora/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBRepository gerencia as operações de leitura/escrita no MongoDB.
type MongoDBRepository struct {
	Client       *mongo.Client
	Database     *mongo.Database
	Ctx          context.Context
	RequestsCol  *mongo.Collection
	ResponsesCol *mongo.Collection
	CountersCol  *mongo.Collection
	UserDataCol  *mongo.Collection
}

// NewMongoDBRepository cria uma nova instância de MongoDBRepository
// com todas as coleções necessárias.
func NewMongoDBRepository(client *mongo.Client, dbName string, ctx context.Context) *MongoDBRepository {
	db := client.Database(dbName)
	return &MongoDBRepository{
		Client:       client,
		Database:     db,
		Ctx:          ctx,
		RequestsCol:  db.Collection("requests"),
		ResponsesCol: db.Collection("responses"),
		CountersCol:  db.Collection("counters"),
		UserDataCol:  db.Collection("user_data"),
	}
}

// SaveUserData salva ou atualiza dados de um usuário (identificado por userNS),
// adicionando os BodyItems ao registro existente, ou criando um novo caso não exista.
func (r *MongoDBRepository) SaveUserData(userNS string, bodyItems []entities.BodyItem, url string) error {
	filter := bson.M{"userNS": userNS}

	update := bson.M{
		"$set": bson.M{
			"userNS": userNS,
			"url":    url,
		},
		"$push": bson.M{
			"body": bson.M{"$each": bodyItems},
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err := r.UserDataCol.UpdateOne(r.Ctx, filter, update, opts); err != nil {
		return fmt.Errorf("erro ao salvar dados do usuário: %v", err)
	}
	return nil
}

// FetchUserData recupera os dados de um usuário a partir de seu userNS.
func (r *MongoDBRepository) FetchUserData(userNS string) (*entities.UserData, error) {
	filter := bson.M{"userNS": userNS}

	var userData entities.UserData
	if err := r.UserDataCol.FindOne(r.Ctx, filter).Decode(&userData); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("Nenhum documento encontrado para userNS: %s", userNS)
		}
		return nil, fmt.Errorf("erro ao buscar dados do usuário: %v", err)
	}
	return &userData, nil
}

// DeleteUserData remove os dados de um usuário do MongoDB após seu envio.
func (r *MongoDBRepository) DeleteUserData(userNS string) error {
	filter := bson.M{"userNS": userNS}
	_, err := r.UserDataCol.DeleteOne(r.Ctx, filter)
	return err
}

// SaveResponseData salva dados relativos a alguma resposta.
func (r *MongoDBRepository) SaveResponseData(responseData interface{}) error {
	_, err := r.ResponsesCol.InsertOne(r.Ctx, responseData)
	return err
}

// IncrementCounters atualiza (ou cria, caso não exista) contadores para um nomeWorkspace e data.
func (r *MongoDBRepository) IncrementCounters(nomeWorkspace string, date string, increments map[string]int) error {
	filter := bson.M{"nomeWorkspace": nomeWorkspace, "contagem.data": date}

	updateFields := bson.M{}
	for field, incValue := range increments {
		updateFields["contagem.$."+field] = incValue
	}

	update := bson.M{"$inc": updateFields}

	result, err := r.CountersCol.UpdateOne(r.Ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar contadores: %v", err)
	}

	// Se não existir, criar um novo documento com os valores default
	if result.MatchedCount == 0 {
		contagemEntry := bson.M{
			"requestRecebidos":    increments["requestRecebidos"],
			"requestEncaminhados": 0,
			"imagensRecebidas":    0,
			"minutos":             0.0,
			"data":                date,
		}

		upsertFilter := bson.M{"nomeWorkspace": nomeWorkspace}
		upsertUpdate := bson.M{"$push": bson.M{"contagem": contagemEntry}}
		if _, err = r.CountersCol.UpdateOne(r.Ctx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("erro ao criar documento de contadores: %v", err)
		}
	}

	return nil
}

// IncrementResponseCounter atualiza (ou cria) a contagem de requestsEncaminhados
// para um nomeWorkspace e data especificados.
func (r *MongoDBRepository) IncrementResponseCounter(nomeWorkspace string, date string, count int) error {
	filter := bson.M{"nomeWorkspace": nomeWorkspace, "contagem.data": date}
	update := bson.M{"$inc": bson.M{"contagem.$.requestEncaminhados": count}}

	result, err := r.CountersCol.UpdateOne(r.Ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar contador de respostas: %v", err)
	}

	if result.MatchedCount == 0 {
		contagemEntry := bson.M{
			"requestRecebidos":    0,
			"requestEncaminhados": count,
			"imagensRecebidas":    0,
			"minutos":             0.0,
			"data":                date,
		}

		upsertFilter := bson.M{"nomeWorkspace": nomeWorkspace}
		upsertUpdate := bson.M{"$push": bson.M{"contagem": contagemEntry}}
		if _, err := r.CountersCol.UpdateOne(r.Ctx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("erro ao criar documento de contadores: %v", err)
		}
	}

	return nil
}

// IncrementImageCounter atualiza (ou cria) a contagem de imagens recebidas.
func (r *MongoDBRepository) IncrementImageCounter(nomeWorkspace string, date string, count int) error {
	filter := bson.M{"nomeWorkspace": nomeWorkspace, "contagem.data": date}
	update := bson.M{"$inc": bson.M{"contagem.$.imagensRecebidas": count}}

	result, err := r.CountersCol.UpdateOne(r.Ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar contador de imagens: %v", err)
	}

	if result.MatchedCount == 0 {
		contagemEntry := bson.M{
			"requestRecebidos":    0,
			"requestEncaminhados": 0,
			"imagensRecebidas":    count,
			"minutos":             0.0,
			"data":                date,
		}

		upsertFilter := bson.M{"nomeWorkspace": nomeWorkspace}
		upsertUpdate := bson.M{"$push": bson.M{"contagem": contagemEntry}}
		if _, err := r.CountersCol.UpdateOne(r.Ctx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("erro ao criar documento de contadores: %v", err)
		}
	}

	return nil
}

// UpdateMinutos atualiza o campo 'minutos' de um registro do workspace e data informados.
func (r *MongoDBRepository) UpdateMinutos(nomeWorkspace string, date string, minutos float64) error {
	filter := bson.M{"nomeWorkspace": nomeWorkspace, "contagem.data": date}
	update := bson.M{"$inc": bson.M{"contagem.$.minutos": minutos}}

	result, err := r.CountersCol.UpdateOne(r.Ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar minutos: %v", err)
	}

	// Se não houver correspondência, criar novo documento com a contagem.
	if result.MatchedCount == 0 {
		contagemEntry := bson.M{
			"requestRecebidos":    0,
			"requestEncaminhados": 0,
			"imagensRecebidas":    0,
			"minutos":             minutos,
			"data":                date,
		}

		upsertFilter := bson.M{"nomeWorkspace": nomeWorkspace}
		upsertUpdate := bson.M{"$push": bson.M{"contagem": contagemEntry}}
		if _, err := r.CountersCol.UpdateOne(r.Ctx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("erro ao criar entrada de contagem: %v", err)
		}
	}

	return nil
}

// IsAccessTokenValid verifica se o access_token fornecido existe na coleção 'security'.
func (r *MongoDBRepository) IsAccessTokenValid(accessToken string) (bool, error) {
	collection := r.Database.Collection("security")
	filter := bson.M{"access_token": accessToken}

	count, err := collection.CountDocuments(r.Ctx, filter)
	if err != nil {
		log.Printf("Erro ao validar access_token: %v", err)
		return false, err
	}

	return count > 0, nil
}
