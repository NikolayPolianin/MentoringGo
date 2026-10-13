package use_case

import (
	"checkgo/OneService/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
)

//go:generate mockgen -source=user_test.go -destination=mocks/mocks.go -package=mocks
type Pgg interface {
	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, id int64) (domain.User, error)

	ReadOutbox(ctx context.Context, limit int) ([]*domain.Event, error)
	SaveOutbox(ctx context.Context, event ...domain.Event) error
	DeleteFromOutbox(ctx context.Context) error
}

//go:generate mockgen -source=user_test.go -destination=mocks/mocks.go -package=mocks
type Publisher interface {
	SendMessage(ctx context.Context, events ...*domain.Event) error
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
	user.Id = userId
	user.UuID = uuid.New()

	eventKey, _ := json.Marshal(user.UuID)
	eventValue, _ := json.Marshal(user)

	event := domain.Event{
		Topic: "user-create",
		Key:   eventKey,
		Value: eventValue,
	}

	if err = u.db.SaveOutbox(ctx, event); err != nil {
		slog.Error("Error saving event", "err", err)
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
