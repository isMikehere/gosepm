package handler

import (
	"testing"
	"time"

	"fmt"

	"../model"
	redis "github.com/go-redis/redis"
)

func TestSetTest(t *testing.T) {

	c := redis.NewClient(&redis.Options{
		Addr:         "sepm:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
	})

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
			// NxTest(tt.args.client)
			// OrderSnGenerator(tt.args.client)

			// ss, err := tt.args.client.HGet("users", "1").Result()

			fmt.Print(ConcatStockList(GetRedisStockCodes(tt.args.client)))
		})
	}
}

func TestGetRedisStock(t *testing.T) {
	c := redis.NewClient(&redis.Options{
		Addr:         "sepm:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
	})

	type args struct {
		client    *redis.Client
		stockCode string
	}
	tests := []struct {
		name string
		args args
		want *model.Stock
	}{
		// TODO: Add test cases.
		{name: "tt", args: args{client: c, stockCode: "000001"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("%s", GetRedisStockDetail(tt.args.client, tt.args.stockCode))
		})
	}
}
