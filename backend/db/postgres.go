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

// NewPostgres creates a new postgres instance.
func NewPostgres(ctx context.Context) (*Postgres, error) {
	postgresurl, _ := ctx.Value(KeyPostgresUrl).(string)
	pool, err := pgxpool.New(ctx, postgresurl)

	// create records table
	_, createTableErr := pool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS records (
			id SERIAL PRIMARY KEY, 
			user_id VARCHAR(36) NOT NULL,
			client_id VARCHAR(36) NOT NULL,
			room_id VARCHAR(36) NOT NULL,
			ts TIMESTAMP NOT NULL DEFAULT current_timestamp,
			duration SERIAL NOT NULL,
			UNIQUE (user_id, client_id, room_id, ts, duration)
		);`,
	)

	// create user column index
	_, createUserIdxErr := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_user ON records(user_id)")
	_, createClientIdxErr := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_client ON records(client_id)")
	_, createRoomIdxErr := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_room ON records(room_id)")
	_, createTsIdxErr := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_ts ON records(ts)")
	_, createDurationIdxErr := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_duration ON records(duration)")

	err = errors.Join(
		err,
		createTableErr,
		createUserIdxErr,
		createClientIdxErr,
		createRoomIdxErr,
		createTsIdxErr,
		createDurationIdxErr,
	)

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
		`SELECT COALESCE(count(*), 0)
		FROM records 
		WHERE duration > (
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
func (p *Postgres) getBestDurationInLeaderboard(
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
