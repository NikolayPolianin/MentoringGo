package postgres

import (
	"checkgo/OneService/internal/domain"
	"checkgo/observability/otel/tracer"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"log/slog"
	"time"
)

type Pg struct {
	*pgxpool.Pool
}

func NewPostgres(ctx context.Context) *Pg {
	addr := fmt.Sprintf("user=postgres password=postgres port=5432 host=localhost dbname=postgres sslmode=disable")
	newcfg, err := pgxpool.ParseConfig(addr)
	if err != nil {
		slog.Error("Error parsing config", "err", err)
	}
	conn, err := pgxpool.NewWithConfig(ctx, newcfg)
	if err != nil {
		slog.Error("Error connecting to database", "err", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		slog.Error("Error pinging database", "err", err)
	}

	return &Pg{conn}
}

func (p *Pg) CreateTable(ctx context.Context) {
	const query = `CREATE TABLE IF NOT EXISTS users (
    Id Bigint generated always as identity PRIMARY KEY,
	Name varchar(255) not null,
	Age int not null check ( Age > 18 ),
	Birthday varchar not null,
	LocalDate date not null)`

	_, err := p.Exec(ctx, query)
	if err != nil {
		slog.Error("Error creating table", "err", err)
		return
	}
}
func (p *Pg) CreateUser(ctx context.Context, user domain.User) (int64, error) {
	ctx, span := tracer.Start(ctx, "Postgres.CreateUser")
	defer span.End()

	var id int64
	query := `INSERT INTO users(Name,Age,Birthday,LocalDate) VALUES ($1,$2,$3,$4) RETURNING Id`

	err := p.QueryRow(ctx, query, user.Name, user.Age, user.Birthday, time.Now()).Scan(&id)
	if err != nil {
		slog.Error("Error creating user", "err", err)
		return 0, err
	}
	span.SetAttributes(attribute.Int64("Id", id))
	return id, nil
}
func (p *Pg) GetUser(ctx context.Context, id int64) (domain.User, error) {
	user := domain.User{}
	query := `SELECT * FROM users WHERE id = $1`

	if err := p.QueryRow(ctx, query, id).Scan(&user.Id, &user.Name, &user.Age, &user.Birthday, &user.LocalDate); err != nil {
		if err == pgx.ErrNoRows {
			slog.Error("User not found", "id", id)
			return user, nil
		}
		slog.Error("Error getting user", "err", err)
	}
	return user, nil
}
