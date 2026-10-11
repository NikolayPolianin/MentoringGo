package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	testDate := domain.User{
		Name:      "Test User",
		Age:       22,
		Birthday:  "2000-01-01",
		LocalDate: time.Now(),
	}
	data, err := json.Marshal(testDate)
	server := NewIntegrationTestServer(t)
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/users", "application/json", bytes.NewReader(data))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var Id int64
	err = json.NewDecoder(resp.Body).Decode(&Id)
	resp, err = http.Get(server.URL + "/api/v1/users/" + strconv.FormatInt(Id, 10))
	if err != nil {
		require.NoError(t, err)
	}
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var user domain.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)
	slog.Info("user Found:", "User: ", user)
	assert.Equal(t, Id, user.Id)

}

func NewIntegrationTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		slog.Error("error creating pool", "err", err)
	}
	require.NoError(t, err)
	err = pool.Ping(ctx)
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "redis",
		DB:       1,
	})
	require.NotNil(t, redisClient)

	producer, err := kafka.NewProducer()
	if err != nil {
		slog.Error("error creating producer", "err", err)
	}
	userRepo := postgres.Pg{pool}
	userCache := redis2.NewRedisClient()
	usecases := use_case.NewUseCase(&userRepo, producer, userCache)
	handler := handler.NewUserHandler(usecases)
	userRouter := http2.NewRouter(handler)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		userRouter.RegisterRoutes(r)
	})

	t.Cleanup(func() {
		redisClient.Close()
		pool.Close()

	})
	return httptest.NewServer(r)
}
