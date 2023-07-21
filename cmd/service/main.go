package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisCtx context.Context

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisCtx = context.Background()

	go func() {
		pubsub := redisClient.Subscribe(redisCtx, "message")
		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(redisCtx)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Channel - %s: %q\n", msg.Channel, msg.Payload)
		}
	}()

	groups, err := redisClient.XInfoGroups(redisCtx, "stream").Result()
	if err != nil {
		fmt.Println(err)
	}

	groupExists := false
	for _, group := range groups {
		if group.Name == "service" {
			groupExists = true
			break
		}
	}

	if !groupExists {
		err := redisClient.XGroupCreateMkStream(redisCtx, "stream", "service", "0").Err()
		if err != nil {
			fmt.Println(err)
		}
	}

	readCount := 0
	for {
		readCount++
		fmt.Println("XREAD number: ", readCount)
		streams, err := redisClient.XReadGroup(redisCtx, &redis.XReadGroupArgs{
			Streams:  []string{"stream", ">"},
			Block:    0,
			Group:    "service",
			Consumer: "worker",
		}).Result()
		if err != nil {
			fmt.Println(err)
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				fmt.Printf("Stream - id: %s - %q\n", message.ID, message.Values)
				redisClient.XAck(redisCtx, "stream", "service", message.ID).Err()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
