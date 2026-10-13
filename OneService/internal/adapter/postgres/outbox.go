package postgres

import (
	"checkgo/OneService/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

func (p *Pg) SaveOutbox(ctx context.Context, event ...domain.Event) error {
	if len(event) == 0 {
		return nil
	}

	sql := `INSERT INTO outbox (topic,key,value,created_at) VALUES ($1,$2,$3,$4)`
	Batch := pgx.Batch{}

	for _, event := range event {
		if event.Topic == "" {
			return domain.ErrTopicIsEmpty
		}
		Batch.Queue(sql, event.Topic, event.Key, event.Value, time.Now())
	}

	result := p.Pool.SendBatch(ctx, &Batch)
	defer result.Close()

	for range Batch.Len() {
		_, err := result.Exec()
		if err != nil {
			return domain.ErrExecOutbox
		}
	}
	return nil
}

func (p *Pg) ReadOutbox(ctx context.Context, limit int) ([]*domain.Event, error) {
	sql := `SELECT topic,key,value FROM outbox
			ORDER BY created_at DESC 
			LIMIT $1`

	events := make([]*domain.Event, 0, limit)
	rows, err := p.Pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.Topic, &event.Key, &event.Value); err != nil {
			return nil, domain.ErrEventIsEmpty
		}
		if rows.Err() != nil {
			rows.Close()
			return nil, errors.New("error scanning rows")
		}
		events = append(events, &event)
	}
	rows.Close()

	return events, nil
}

func (p *Pg) DeleteFromOutbox(ctx context.Context) error {
	sql := `TRUNCATE outbox`
	_, err := p.Pool.Exec(ctx, sql)
	if err != nil {
		return err
	}
	return nil
}
