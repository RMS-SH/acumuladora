// ==================================================
// infrastructure/queue/redis_queue.go
// ==================================================
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RMS-SH/filaFlowise/domain"
	"github.com/redis/go-redis/v9"
)

// RedisQueue implementação da interface domain.Queue usando Redis
type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisQueueFromClient cria uma nova instância de RedisQueue com um cliente Redis existente
func NewRedisQueueFromClient(client *redis.Client) (*RedisQueue, error) {
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Redis: %v", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

// Enqueue adiciona um item à fila
func (q *RedisQueue) Enqueue(userNs string, item domain.QueueItem) error {
	key := fmt.Sprintf("queue:%s", userNs)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return q.client.RPush(q.ctx, key, data).Err()
}

// Dequeue remove um item da fila
func (q *RedisQueue) Dequeue(userNs string) (domain.QueueItem, error) {
	key := fmt.Sprintf("queue:%s", userNs)
	data, err := q.client.LPop(q.ctx, key).Result()
	if err != nil {
		return domain.QueueItem{}, err
	}

	var item domain.QueueItem
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return domain.QueueItem{}, err
	}
	return item, nil
}

// TryLock tenta adquirir um lock para o userNs com um TTL
func (q *RedisQueue) TryLock(userNs string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:%s", userNs)
	ok, err := q.client.SetNX(q.ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// ReleaseLock libera o lock
func (q *RedisQueue) ReleaseLock(userNs string) error {
	key := fmt.Sprintf("lock:%s", userNs)
	return q.client.Del(q.ctx, key).Err()
}

// IsLocked verifica se o lock existe
func (q *RedisQueue) IsLocked(userNs string) (bool, error) {
	key := fmt.Sprintf("lock:%s", userNs)
	res, err := q.client.Exists(q.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
