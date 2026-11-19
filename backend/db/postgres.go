package db

import (
	"context"
	"errors"
	"time"

	"buttonmania.win/protocol"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres represents the postgres client.
type Postgres struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

// schemaStatements contains idempotent statements to create and migrate schema.
var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS records (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL,
		client_id VARCHAR(36) NOT NULL,
		room_id VARCHAR(36) NOT NULL,
		ts TIMESTAMP NOT NULL DEFAULT current_timestamp,
		duration BIGINT NOT NULL,
		UNIQUE (user_id, client_id, room_id, ts, duration)
	);`,
	// migrate duration from SERIAL, which caps values at int32
	`DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'records'
			AND column_name = 'duration'
			AND data_type <> 'bigint'
		) THEN
			ALTER TABLE records ALTER COLUMN duration DROP DEFAULT;
			ALTER TABLE records ALTER COLUMN duration TYPE BIGINT;
		END IF;
	END $$;`,
	`DROP SEQUENCE IF EXISTS records_duration_seq;`,
	// composite indexes, leaderboard queries always filter by client and room
	`CREATE INDEX IF NOT EXISTS idx_records_room_duration
		ON records(client_id, room_id, duration DESC);`,
	`CREATE INDEX IF NOT EXISTS idx_records_room_ts
		ON records(client_id, room_id, ts DESC);`,
	`CREATE INDEX IF NOT EXISTS idx_records_room_user
		ON records(client_id, room_id, user_id, duration DESC);`,
	// drop single column indexes, superseded by composite ones
	`DROP INDEX IF EXISTS idx_user;`,
	`DROP INDEX IF EXISTS idx_client;`,
	`DROP INDEX IF EXISTS idx_room;`,
	`DROP INDEX IF EXISTS idx_ts;`,
	`DROP INDEX IF EXISTS idx_duration;`,
}

// NewPostgres creates a new postgres instance.
func NewPostgres(ctx context.Context) (*Postgres, error) {
	postgresurl, _ := ctx.Value(KeyPostgresUrl).(string)
	pool, err := pgxpool.New(ctx, postgresurl)
	if err != nil {
		return nil, err
	}

	for _, statement := range schemaStatements {
		if _, execErr := pool.Exec(ctx, statement); execErr != nil {
			err = errors.Join(err, execErr)
		}
	}

	return &Postgres{
		ctx:  ctx,
		pool: pool,
	}, err
}

// closes the postgres connection.
func (p *Postgres) close() error {
	defer p.pool.Close()
	return nil
}

// adds a gameplay record to the leaderboard.
func (p *Postgres) addRecordToLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
	record protocol.GameplayRecord,
) error {
	_, err := p.pool.Exec(
		p.ctx,
		`INSERT INTO records(user_id, client_id, room_id, ts, duration) 
		VALUES($1, $2, $3, $4, $5) 
		ON CONFLICT DO NOTHING`,
		userID,
		clientId,
		roomId,
		time.Unix(record.Timestamp, 0),
		record.Duration,
	)
	return err
}

// retrieves the duration place in the leaderboard.
func (p *Postgres) getDurationPlaceInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	duration int64,
) (int64, error) {
	var count int64
	err := p.pool.QueryRow(
		p.ctx,
		`SELECT COALESCE(count(DISTINCT duration), 0) 
		FROM records 
		WHERE client_id=$1 AND room_id=$2 AND duration > $3`,
		clientId,
		roomId,
		duration,
	).Scan(&count)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// retrieves the user's place in the leaderboard.
func (p *Postgres) getUserPlaceInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
	userID protocol.UserID,
) (int64, error) {
	var count int64
	err := p.pool.QueryRow(
		p.ctx,
		`SELECT COALESCE(count(DISTINCT user_id), 0)
		FROM records
		WHERE client_id=$1 AND room_id=$2 AND duration > (
			SELECT COALESCE(MAX(duration), 0)
			FROM records
			WHERE client_id=$1 AND room_id=$2 AND user_id=$3
		)`,
		clientId,
		roomId,
		userID,
	).Scan(&count)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// retrieves the count of users in the leaderboard.
func (p *Postgres) getUsersCountInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	var count int64
	err := p.pool.QueryRow(
		p.ctx,
		`SELECT COALESCE(count(DISTINCT user_id), 0)
		FROM records 
		WHERE client_id=$1 AND room_id=$2 AND duration > 0`,
		clientId,
		roomId,
	).Scan(&count)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// retrieves the best duration achieved by a player in the leaderboard.
func (p *Postgres) getBestOverallDurationInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	var duration int64
	err := p.pool.QueryRow(
		p.ctx,
		`SELECT COALESCE(MAX(duration), 0)
		FROM records 
		WHERE client_id=$1 AND room_id=$2 AND duration > 0`,
		clientId,
		roomId,
	).Scan(&duration)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return duration, err
}

// retrieves today's best duration from the leaderboard.
func (p *Postgres) getTodaysDurationInLeaderboard(
	clientId protocol.ClientID,
	roomId protocol.RoomID,
) (int64, error) {
	var duration int64
	err := p.pool.QueryRow(
		p.ctx,
		`SELECT COALESCE(MAX(duration), 0)
		FROM records 
		WHERE client_id=$1 AND room_id=$2 AND ts >= now()::date AND duration > 0`,
		clientId,
		roomId,
	).Scan(&duration)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return duration, err
}
