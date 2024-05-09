package main

import (
	"context"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"unreal.sh/echo/internal/server"
)

func main() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync()
	logger := baseLogger.Sugar()

	godotenv.Load(".env")

	ctx := context.Background()
	server.Start(ctx, logger)
}
