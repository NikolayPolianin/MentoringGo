package handler

import (
	"checkgo/OneService/internal/domain"
	"checkgo/OneService/internal/infrastructure/jwt"
	"checkgo/OneService/internal/transport/http/response"
	"checkgo/observability/otel/tracer"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"log/slog"
	"net/http"
	"strconv"
)

type UseCase interface {
	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
}
type UserHandler struct {
	u UseCase
}

func NewUserHandler(useCase UseCase) *UserHandler {
	return &UserHandler{u: useCase}
}

// CreateUser
// @Summary Создать Юзера
// @Description Создать Альберта
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.User true "User"
// @Success 200 {object} domain.User
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handler.CreateUser")
	defer span.End()
	var user domain.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Warn("Error Parsing userBody: ", r.Body)
	}
	span.SetAttributes(attribute.String("name", user.Name), attribute.Int("age", user.Age))
	id, err := h.u.CreateUser(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Error Creating User: ", r.Body, "traceId:", span.SpanContext().TraceID().String())
		response.Error(w, 400, "Error in CreateUser")
		return
	}
	response.JSON(w, http.StatusOK, id)
	slog.Info("create user success")
}

// GetUser
// @Summary Получить Юзера
// @Description Создать Альберта
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @in header
// @name Authorization
// @Param id path int true "ID пользователя"
// @Success 200 {object} domain.User
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	Id := chi.URLParam(r, "id")

	user, err := h.u.GetUser(r.Context(), Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		slog.Warn("Error in GetUser: ", r.Body)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
	}
	slog.Info("get user success")
	response.JSON(w, http.StatusOK, user)
}

// CreateAutToken
// @Summary Создать Токен
// @Description Создать Токен
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /api/v1/jwt/{id} [post]
func (h *UserHandler) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	Id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		slog.Warn("Error in CreateAuthToken: ", r.URL.Query().Get("id"))
		response.Error(w, http.StatusBadRequest, "Error in Parsing Id")
		return
	}

	token, err := jwt.CreateToken(Id, "admin", []byte("My_secret"))
	if err != nil {
		slog.Warn("Error in CreateAuthToken: for Id:", r.URL.Query().Get("id"))
		response.Error(w, http.StatusBadRequest, "Error in CreateAuthToken")
		return
	}

	slog.Info("create auth token success")
	response.JSON(w, http.StatusOK, token)
}
