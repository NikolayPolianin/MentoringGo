package use_case

import (
	"checkgo/OneService/internal/domain"
	"checkgo/OneService/internal/use_case/mocks"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
	"time"
)
import "go.uber.org/mock/gomock"

func TestUserUseCase_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockPgg(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	user := domain.User{
		Id:        10,
		Name:      "Test User",
		Age:       23,
		Birthday:  "2025-01-01",
		LocalDate: time.Now(),
	}
	mockCache.EXPECT().
		GetUser(gomock.Any(), "10").
		Return(nil, redis.Nil)
	mockDB.EXPECT().
		GetUser(gomock.Any(), int64(10)).
		Return(user, nil)
	mockCache.EXPECT().
		SetUser(gomock.Any(), "10", &user).
		Return(nil)

	uc := NewUseCase(
		mockDB,
		mockPublisher,
		mockCache,
	)

	result, err := uc.GetUser(context.Background(), "10")
	assert.NoError(t, err)
	require.Equal(t, user, result)
}

func TestUserUseCase_GetUser_From_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockPgg(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	user := domain.User{
		Id:        10,
		Name:      "Test User",
		Age:       23,
		Birthday:  "2025-01-01",
		LocalDate: time.Now(),
	}
	mockCache.EXPECT().
		GetUser(gomock.Any(), "10").
		Return(user, nil)

	uc := NewUseCase(
		mockDB,
		mockPublisher,
		mockCache,
	)
	result, err := uc.GetUser(context.Background(), "10")
	require.NoError(t, err)
	require.Equal(t, user, result)
}

func TestUserValidate(t *testing.T) {
	user := domain.User{
		Id:        10,
		Name:      "Valid Test User",
		Age:       23,
		Birthday:  "2025-01-01",
		LocalDate: time.Now(),
	}
	badUser := domain.User{
		Id:        10,
		Name:      "",
		Age:       14,
		Birthday:  "2025-01-01",
		LocalDate: time.Now(),
	}
	cases := []struct {
		user  domain.User
		valid bool
	}{
		{user: user,
			valid: true},
		{badUser, false},
	}
	for _, tt := range cases {
		t.Run(tt.user.Name, func(t *testing.T) {
			err := domain.ValidateUser(&tt.user)
			if err != nil {
				slog.Error("User validation failed for ", "errors:")
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.user.Name, tt.user.Name)
		})
	}
}
