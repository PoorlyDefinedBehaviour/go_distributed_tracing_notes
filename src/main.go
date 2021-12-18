package main

import (
	"context"
	"fmt"
	"kata/src/logger"
	"kata/src/reqwest"

	"github.com/google/uuid"
)

// go test -coverprofile=count.out ./... && go tool cover -html=count.out -o coverage.html
func main() {
	log := logger.New()

	log.Prefix(fmt.Sprintf("project-name::x-rvbr-request-id=%s", uuid.New().String()))

	log.Info("test")

	type User struct {
	}

	repos := make([]map[string]interface{}, 0)

	err := reqwest.GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
		Header("key", "value").
		Build().
		JSON(&repos)
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("got %d repos", len(repos)))
}

// A logs using log.Info() -> A::x-rvbr-request-id=d39e1004-f2ae-4883-9ead-2926dd435e26
// A request B
// B logs using log.Info() -> B::x-rvbr-request-id=d39e1004-f2ae-4883-9ead-2926dd435e26

// Use context
// (Request, Context)
// -> Middleware
// -> (Request, Context with correlation id)
// -> handler

// package http
//
// func GET(ctx context.Context, url string) {
// 		defer log.Trace(ctx, "GET %s", url)()
//    ...
// 		correlationID := ctx.Get("x-correlation-id")
//    make http request with correlation id
//    ...
// }
