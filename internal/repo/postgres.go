package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pavel97go/subscriptions/internal/domain"
	"github.com/pavel97go/subscriptions/internal/logger"
	"github.com/pavel97go/subscriptions/internal/util"
)

type Repo struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Repo, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	r := &Repo{db: pool}
	if err := r.applyMigrations(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return r, nil
}

func (r *Repo) Close() { r.db.Close() }

func (r *Repo) applyMigrations(ctx context.Context) error {
	logger.Log.Info("applying migrations...")
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "./migrations"
	}
	path := filepath.Join(dir, "001_init.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}
	if _, err := r.db.Exec(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	logger.Log.Info("migrations applied successfully")
	return nil
}

func (r *Repo) Create(ctx context.Context, s domain.Subscription) (uuid.UUID, error) {
	logger.Log.Infof("creating subscription: user_id=%s, service=%s", s.UserID, s.ServiceName)
	id := uuid.New()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.db.Exec(ctx, `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_month, end_month)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		id, s.ServiceName, s.Price, s.UserID, s.StartMonth, s.EndMonth,
	)
	if err != nil {
		logger.Log.Errorf("create exec error: %v", err)
	}
	return id, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var s domain.Subscription
	var end *time.Time
	err := r.db.QueryRow(ctx, `
		SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
		  FROM subscriptions WHERE id=$1`, id).
		Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartMonth, &end, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		logger.Log.Errorf("get query error: %v", err)
		return s, err
	}
	s.EndMonth = end
	return s, nil
}

func (r *Repo) List(ctx context.Context, limit, offset int) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := r.db.Query(ctx, `
		SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
		  FROM subscriptions
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		logger.Log.Errorf("list query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var out []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var end *time.Time
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartMonth, &end, &s.CreatedAt, &s.UpdatedAt); err != nil {
			logger.Log.Errorf("list scan error: %v", err)
			return nil, err
		}
		s.EndMonth = end
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("list rows error: %v", err)
		return nil, err
	}
	return out, nil
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, s domain.Subscription) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.db.Exec(ctx, `
		UPDATE subscriptions
		   SET service_name=$2, price=$3, user_id=$4, start_month=$5, end_month=$6, updated_at=now()
		 WHERE id=$1`,
		id, s.ServiceName, s.Price, s.UserID, s.StartMonth, s.EndMonth,
	)
	if err != nil {
		logger.Log.Errorf("update exec error: %v", err)
	}
	return err
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.db.Exec(ctx, `DELETE FROM subscriptions WHERE id=$1`, id)
	if err != nil {
		logger.Log.Errorf("delete exec error: %v", err)
	}
	return err
}

type SummaryFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	From, To    time.Time
}

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
}

func (r *Repo) ListFiltered(ctx context.Context, f ListFilter, limit, offset int) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var args []any
	var conds []string
	i := 1
	if f.UserID != nil {
		conds = append(conds, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *f.UserID)
		i++
	}
	if f.ServiceName != nil {
		conds = append(conds, fmt.Sprintf("service_name = $%d", i))
		args = append(args, *f.ServiceName)
		i++
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	q := fmt.Sprintf(`
		SELECT id, service_name, price, user_id, start_month, end_month, created_at, updated_at
		  FROM subscriptions
		  %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`, where, i, i+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		logger.Log.Errorf("list filtered query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var out []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var end *time.Time
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartMonth, &end, &s.CreatedAt, &s.UpdatedAt); err != nil {
			logger.Log.Errorf("list filtered scan error: %v", err)
			return nil, err
		}
		s.EndMonth = end
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("list filtered rows error: %v", err)
		return nil, err
	}
	return out, nil
}

func (r *Repo) Summary(ctx context.Context, f SummaryFilter) (int, error) {
	logger.Log.Infof("summary requested: from=%v to=%v user_id=%v service_name=%v", f.From, f.To, f.UserID, f.ServiceName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var args []any
	var conds []string
	conds = append(conds, `NOT (end_month IS NOT NULL AND end_month < $1) AND start_month <= $2`)
	args = append(args, f.From, f.To)

	i := 3
	if f.UserID != nil {
		conds = append(conds, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *f.UserID)
		i++
	}
	if f.ServiceName != nil {
		conds = append(conds, fmt.Sprintf("service_name = $%d", i))
		args = append(args, *f.ServiceName)
		i++
	}

	q := `SELECT service_name, price, user_id, start_month, end_month
	        FROM subscriptions
	       WHERE ` + strings.Join(conds, " AND ")

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		logger.Log.Errorf("summary query error: %v", err)
		return 0, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var name string
		var price int
		var uid uuid.UUID
		var start time.Time
		var end *time.Time
		if err := rows.Scan(&name, &price, &uid, &start, &end); err != nil {
			logger.Log.Errorf("summary scan error: %v", err)
			return 0, err
		}
		months := util.MonthsOverlap(start, end, f.From, f.To)
		total += price * months
	}
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("summary rows error: %v", err)
		return 0, err
	}
	logger.Log.Infof("summary total=%d", total)
	return total, nil
}
