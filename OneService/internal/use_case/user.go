package use_case

import (
	"checkgo/OneService/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"time"
)

//go:generate mockgen -source=user_test.go -destination=mocks/mocks.go -package=mocks
type Pgg interface {
	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, id int64) (domain.User, error)
}

//go:generate mockgen -source=user_test.go -destination=mocks/mocks.go -package=mocks
type Publisher interface {
	SendMessage(message *sarama.ProducerMessage) error
}

//go:generate mockgen -source=user_test.go -destination=mocks/mocks.go -package=mocks
type Cache interface {
	GetUser(ctx context.Context, id string) (*domain.User, error)
	SetUser(ctx context.Context, id string, user *domain.User) error
}
type User struct {
	db       Pgg
	producer Publisher
	cache    Cache
}

func NewUseCase(db Pgg, producer Publisher, cache Cache) *User {
	return &User{db, producer, cache}
}

func (u *User) CreateUser(ctx context.Context, user domain.User) (int64, error) {
	userId, err := u.db.CreateUser(ctx, user)
	if err != nil {
		slog.Error("Error creating user", "err", err)
		return -1, err
	}
	valueMsg, err := json.Marshal(user)
	if err != nil {
		slog.Error("Error marshalling user", "err", err)
		return -1, err
	}

	msg := &sarama.ProducerMessage{
		Topic:     "user-create",
		Key:       sarama.StringEncoder(strconv.Itoa(int(userId))),
		Value:     sarama.ByteEncoder(valueMsg),
		Timestamp: time.Now(),
	}
	if err := u.producer.SendMessage(msg); err != nil {
		slog.Error("Error sending message", "err", err)
		return -1, err
	}
	return userId, nil
}

func (u *User) GetUser(ctx context.Context, id string) (domain.User, error) {
	cacheUser, err := u.cache.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			slog.Error("Error in Cache", "error:", err)
		}
		slog.Warn("GetUser err:", "error:", err)
	}
	if cacheUser != nil {
		slog.Info("GetUser From Cache:",
			"UserId", cacheUser.Id,
			"Name", cacheUser.Name,
			"Birthday", cacheUser.Birthday,
			"LocalDate", cacheUser.LocalDate)
		return *cacheUser, nil
	}
	ParseId, err := strconv.ParseInt(id, 10, 64)
	user, err := u.db.GetUser(ctx, ParseId)
	if err != nil {
		slog.Error("Error getting user", "err", err)
		return domain.User{}, err
	}
	if err := u.cache.SetUser(ctx, id, &user); err != nil {
		slog.Warn("SetUser:", "err", err)
		return user, err
	}
	return user, nil
}
