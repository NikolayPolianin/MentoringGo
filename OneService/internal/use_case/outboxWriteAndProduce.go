package use_case

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func (uc *User) OutboxReadAndProduce(ctx context.Context, limit int) (int, error) {
	var count int

	if err := uc.tx.WithTransaction(ctx, func(tx pgx.Tx) error {
		events, err := uc.db.ReadOutbox(ctx, tx, limit)
		if err != nil {
			return err
		}

		count := len(events)
		if count < limit {
			return nil
		}

		if err := uc.producer.SendMessage(ctx, events...); err != nil {
			return err
		}

		err = uc.db.DeleteFromOutbox(ctx, tx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return 0, err
	}
	return count, nil
}
