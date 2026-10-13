package use_case

import (
	"context"
)

func (uc *User) OutboxReadAndProduce(ctx context.Context, limit int) (int, error) {

	events, err := uc.db.ReadOutbox(ctx, limit)
	if err != nil {
		return 0, err
	}
	count := len(events)
	if count < limit {
		return count, nil
	}
	if err := uc.producer.SendMessage(ctx, events...); err != nil {
		return 0, err
	}
	err = uc.db.DeleteFromOutbox(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}
