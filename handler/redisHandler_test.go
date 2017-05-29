package handler

import (
	"testing"
	"time"

	redis "github.com/go-redis/redis"
)

func TestSetTest(t *testing.T) {

	c := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
	})
	c.FlushDb()

	type args struct {
		client *redis.Client
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{name: "tt", args: args{client: c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTest(tt.args.client)
		})
	}
}
