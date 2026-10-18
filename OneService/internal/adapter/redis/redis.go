package redis

import (
	"checkgo/OneService/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedisClient() *Redis {
	return &Redis{client: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       1,
	})}
}
func (r *Redis) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	if err := r.client.Close(); err != nil {
		slog.Error("error closing redis client:", err)
	}
}

func (p *Redis) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := p.client.Ping(ctx)
	if err := status.Err(); err != nil {
		slog.Error("Error pinging Redis", err)
		return err
	}
	return nil
}
func (p *Redis) GetUser(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	cachedUser, err := p.client.Get(ctx, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
		return nil, err
	}
	if err := json.Unmarshal(cachedUser, &user); err != nil {
		return nil, err
	}

	if user.Id == 0 {
		p.client.Del(ctx, id)
		return nil, nil
	}
	slog.Info("User found in Cache:", "id", id)
	return &user, nil
}
func (p *Redis) SetUser(ctx context.Context, id string, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	if err := p.client.Set(ctx, id, data, 5*time.Minute).Err(); err != nil {
		slog.Error("Error setting user", err)
		return err
	}
	return nil
}
