package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexMickh/proj-user/internal/app"
	"github.com/AlexMickh/proj-user/internal/config"
	"github.com/AlexMickh/proj-user/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Hour)
	defer cancel()

	ctx = logger.New(ctx, []string{"stdout"}, cfg.Env)

	logger.FromCtx(ctx).Info("logger is working", zap.String("env", cfg.Env))

	app := app.Register(ctx, cfg)
	app.Run(ctx)
	defer app.GracefulStop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	close(stop)
	logger.FromCtx(ctx).Info("server stopped")
}
