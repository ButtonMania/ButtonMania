//go:build integration

package db

import (
	"context"
	"errors"
	"os"
)

// IntegrationContext builds a db context from the integration environment.
func IntegrationContext() context.Context {
	postgresUrl := os.Getenv("TEST_POSTGRES_URL")
	if postgresUrl == "" {
		postgresUrl = "postgres://postgres:password@localhost:55432/buttonmania_test"
	}
	redisAddress := os.Getenv("TEST_REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:56379"
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPostgresUrl, postgresUrl)
	ctx = context.WithValue(ctx, KeyRedisAddress, redisAddress)
	ctx = context.WithValue(ctx, KeyRedisUsername, "")
	ctx = context.WithValue(ctx, KeyRedisPassword, "")
	ctx = context.WithValue(ctx, KeyRedisDatabase, 0)
	ctx = context.WithValue(ctx, KeyRedisTLS, false)
	return ctx
}

// FlushForTests drops all data from both databases.
func (db *DB) FlushForTests() error {
	flushRedisErr := db.redis.client.FlushDB(db.redis.ctx).Err()
	_, truncateErr := db.postgres.pool.Exec(db.postgres.ctx, "TRUNCATE records")
	return errors.Join(flushRedisErr, truncateErr)
}
