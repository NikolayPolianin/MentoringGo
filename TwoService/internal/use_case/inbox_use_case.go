package use_case

import (
	"checkgo/TwoService/internal/domain"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"log/slog"
)

type InboxRepository interface {
	WriteInbox(ctx context.Context, users ...domain.User) error
}
type InboxUseCase struct {
	UserRepository InboxRepository
}

func NewInboxUseCase(userRepository InboxRepository) *InboxUseCase {
	return &InboxUseCase{userRepository}
}

func (u *InboxUseCase) ReadAndWriteInbox(ctx context.Context, events ...*sarama.ConsumerMessage) error {
	var users []domain.User
	for _, event := range events {
		var user domain.User
		if err := json.Unmarshal(event.Value, &user); err != nil {
			slog.Error("error unmarshalling inbox user", err)
		}
		if err := json.Unmarshal(event.Key, &user.UuID); err != nil {
			slog.Error("error unmarshalling inbox user", err)
		}
		if err := u.UserRepository.WriteInbox(ctx, user); err != nil {
			slog.Error("error writing inbox user", err)
		}
		users = append(users, user)
	}
	if err := u.UserRepository.WriteInbox(ctx, users...); err != nil {
		slog.Error("error writing inbox user", "err", err)
	}
	return nil
}
