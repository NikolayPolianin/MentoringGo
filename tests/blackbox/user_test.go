package blackbox

import (
	"checkgo/OneService/internal/domain"
	http2 "checkgo/OneService/internal/transport/http"
	"checkgo/OneService/internal/transport/http/handler"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockUseCase struct {
	GetUserFunc func(ctx context.Context, id string) (domain.User, error)
}

func (m *MockUseCase) CreateUser(
	ctx context.Context,
	user domain.User,
) (int64, error) {
	return 0, nil
}

func (m *MockUseCase) GetUser(
	ctx context.Context,
	id string,
) (domain.User, error) {
	return m.GetUserFunc(ctx, id)
}

type User struct {
	Id        int64     `json:"id,omitempty"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Birthday  string    `json:"birthday"`
	LocalDate time.Time `json:"localDate"`
}

func TestGetUserApi(t *testing.T) {
	mockUseCase := &MockUseCase{
		GetUserFunc: func(ctx context.Context, id string) (domain.User, error) {
			require.Equal(t, "26", id)
			return domain.User{
				Id: 26,
			}, nil
		},
	}
	handler := handler.NewUserHandler(mockUseCase)
	apiRouter := http2.NewRouter(handler)
	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		apiRouter.RegisterRoutes(r)
	})
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/users/26",
		nil,
	)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var user domain.User
	err := json.Unmarshal(rec.Body.Bytes(), &user)
	slog.Info("get user", "user:", user)
	require.NoError(t, err)
	assert.Equal(t, int64(26), user.Id)
}
