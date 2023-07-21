package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisCtx context.Context

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisCtx = context.Background()

	r := http.NewServeMux()

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/pub-sub", pubsubHandler)
	r.HandleFunc("/stream", streamHandler)

	fmt.Println("Listening on port 4000")
	http.ListenAndServe(":4000", r)
}

func pubsubHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.FormValue("message")
	if msg == "" {
		http.Redirect(w, r, "/", http.StatusBadRequest)
	}

	err := redisClient.Publish(redisCtx, "message", msg).Err()
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.FormValue("message")
	if msg == "" {
		http.Redirect(w, r, "/", http.StatusBadRequest)
	}

	err := redisClient.XAdd(redisCtx, &redis.XAddArgs{
		Stream: "stream",
		Values: map[string]any{
			"message": msg,
		},
		ID: "*",
	}).Err()
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <title>Redis-Go-Playground</title>
        </head>
        <body>
            <h1>Pub/Sub</h1>
            <form action="/pub-sub" method="post">
                <input name="message" placeholder="Message..." />
                <input type="submit" />
            </form>
            <h1>Stream</h1>
            <form action="/stream" method="post">
                <input name="message" placeholder="Message..." />
                <input type="submit" />
            </form>
        </body>
        </html>
        `))
}
