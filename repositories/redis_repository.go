////////////////////////////////////////////////////////////////////////////////
// repositories/redis_repository.go
////////////////////////////////////////////////////////////////////////////////

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RMS-SH/acumuladora/entities"
	"github.com/go-redis/redis/v8"
)

// RedisRepository gerencia as operações de leitura/escrita no Redis.
type RedisRepository struct {
	Client *redis.Client
}

// NewRedisRepository cria uma nova instância de RedisRepository.
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		Client: client,
	}
}

// SaveUserData salva ou atualiza dados de um usuário (identificado por userNS) no Redis.
// O valor é armazenado como JSON no key "user_data:{userNS}".
func (r *RedisRepository) SaveUserData(userNS string, bodyItems []entities.BodyItem, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("user_data:%s", userNS)

	// Buscar se já existe algo salvo
	existing, err := r.Client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("erro ao obter dados do Redis: %v", err)
	}

	var userData entities.UserData
	// Se existir, decodificar e mesclar
	if existing != "" {
		if err := json.Unmarshal([]byte(existing), &userData); err != nil {
			return fmt.Errorf("erro ao desserializar dados existentes: %v", err)
		}
	} else {
		userData = entities.UserData{}
	}

	userData.UserNS = userNS
	userData.URL = url
	userData.Body = append(userData.Body, bodyItems...)

	// Serializar e armazenar
	dataBytes, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("erro ao serializar dados do usuário: %v", err)
	}
	if err := r.Client.Set(ctx, key, dataBytes, 0).Err(); err != nil {
		return fmt.Errorf("erro ao salvar dados no Redis: %v", err)
	}

	return nil
}

// FetchUserData recupera os dados de um usuário a partir de seu userNS, retornando entities.UserData.
func (r *RedisRepository) FetchUserData(userNS string) (*entities.UserData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("user_data:%s", userNS)
	result, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("nenhum dado encontrado para userNS: %s", userNS)
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao obter dados do Redis: %v", err)
	}

	var userData entities.UserData
	if err := json.Unmarshal([]byte(result), &userData); err != nil {
		return nil, fmt.Errorf("erro ao desserializar dados do usuário: %v", err)
	}

	return &userData, nil
}

// DeleteUserData remove os dados de um usuário do Redis após seu envio.
func (r *RedisRepository) DeleteUserData(userNS string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("user_data:%s", userNS)
	if err := r.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("erro ao deletar dados do usuário no Redis: %v", err)
	}
	return nil
}
