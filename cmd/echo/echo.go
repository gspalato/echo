package main

import (
	"context"

	"github.com/joho/godotenv"

	"unreal.sh/echo/internal/server"
)

func main() {
	godotenv.Load(".env")

	ctx := context.Background()
	server.Start(ctx)
}
