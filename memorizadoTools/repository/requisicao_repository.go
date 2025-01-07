package repository

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

// RequisicaoRepository define interface para acessar dados de "Requisicao" no Redis
type RequisicaoRepository interface {
	ObterDados(ctx context.Context, compositeKey string) (map[string]interface{}, time.Time, error)
	SalvarDados(ctx context.Context, compositeKey string, dados map[string]interface{}, expiracao time.Duration) error
}

type requisicaoRepository struct {
	redisClient *redis.Client
}

// NovoRequisicaoRepository cria uma instância do repositório
func NovoRequisicaoRepository(redisClient *redis.Client) RequisicaoRepository {
	log.Println("Iniciando RequisicaoRepository...")
	return &requisicaoRepository{
		redisClient: redisClient,
	}
}

// ObterDados retorna dados já armazenados e o timestamp salvo, se houver
func (r *requisicaoRepository) ObterDados(ctx context.Context, compositeKey string) (map[string]interface{}, time.Time, error) {
	log.Printf("Buscando dados no Redis para compositeKey=%s", compositeKey)
	res, err := r.redisClient.HGetAll(ctx, compositeKey).Result()
	if err != nil {
		log.Printf("Erro ao buscar HGetAll no Redis: %v", err)
		return nil, time.Time{}, err
	}
	if len(res) == 0 {
		log.Printf("Nenhum dado encontrado no Redis para compositeKey=%s", compositeKey)
		return nil, time.Time{}, nil
	}

	dadosJSON, ok := res["dados"]
	if !ok {
		log.Printf("Campo 'dados' não encontrado em compositeKey=%s", compositeKey)
		return nil, time.Time{}, nil
	}

	timestampStr, ok := res["timestamp"]
	if !ok {
		log.Printf("Campo 'timestamp' não encontrado em compositeKey=%s", compositeKey)
		return nil, time.Time{}, nil
	}

	var dados map[string]interface{}
	if err := json.Unmarshal([]byte(dadosJSON), &dados); err != nil {
		log.Printf("Erro ao deserializar dados JSON: %v", err)
		return nil, time.Time{}, err
	}

	tsInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		log.Printf("Erro ao converter timestamp: %v", err)
		return nil, time.Time{}, err
	}
	timestamp := time.Unix(tsInt, 0)

	log.Printf("Dados encontrados no Redis compositeKey=%s => %v, timestamp=%v", compositeKey, dados, timestamp)
	return dados, timestamp, nil
}

// SalvarDados armazena dados e um timestamp de atualização no Redis, com expiracao
func (r *requisicaoRepository) SalvarDados(ctx context.Context, compositeKey string, dados map[string]interface{}, expiracao time.Duration) error {
	log.Printf("Salvando dados no Redis para compositeKey=%s com TTL=%v", compositeKey, expiracao)
	ID := dados["id"]
	delete(dados, "id")

	bytesJSON, err := json.Marshal(dados)
	if err != nil {
		log.Printf("Erro ao serializar dados: %v", err)
		return err
	}

	timestamp := time.Now().Unix()

	_, err = r.redisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, compositeKey, "dados", bytesJSON)
		pipe.HSet(ctx, compositeKey, "timestamp", timestamp)
		pipe.Expire(ctx, compositeKey, expiracao)
		return nil
	})

	// Restaura o "id" no map, se necessário
	dados["id"] = ID

	if err != nil {
		log.Printf("Erro ao salvar no Redis para compositeKey=%s: %v", compositeKey, err)
		return err
	}
	log.Printf("Dados salvos com sucesso no Redis para compositeKey=%s", compositeKey)
	return nil
}
