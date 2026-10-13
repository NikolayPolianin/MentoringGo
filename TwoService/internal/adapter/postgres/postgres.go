package postgres

import (
	"checkgo/TwoService/internal/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type Postgres struct {
	*pgxpool.Pool
}

func NewPostgres(ctx context.Context) *Postgres {
	addr := fmt.Sprintf("user=postgres password=postgres port=5432 host=localhost dbname=postgres sslmode=disable")
	cfg, err := pgxpool.ParseConfig(addr)
	if err != nil {
		slog.Error("error parsing config", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("error connecting to postgres", err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctxTimeout); err != nil {
		slog.Error("error pinging postgres", err)
	}

	return &Postgres{conn}
}

func (p *Postgres) WriteInbox(ctx context.Context, users ...domain.User) error {
	sql := `INSERT INTO inbox (uuid,name,age,birthday,localdate,created_at) VALUES ($1, $2, $3, $4, $5,$6)
		ON CONFLICT (uuid) DO NOTHING`

	if len(users) == 0 {
		return nil
	}

	for _, user := range users {
		if _, err := p.Exec(ctx, sql, user.UuID, user.Name, user.Age, user.Birthday, user.LocalDate, time.Now()); err != nil {
			slog.Error(
				"failed to insert inbox user",
				"uuid", user.UuID,
				"error", err,
			)
			continue
		}
	}

	return nil
}
