package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/nulln0ne/uniswap-estimator/internal/config"
	"github.com/nulln0ne/uniswap-estimator/internal/eth"
	"github.com/nulln0ne/uniswap-estimator/internal/handler"
	"github.com/nulln0ne/uniswap-estimator/internal/logging"
	"github.com/nulln0ne/uniswap-estimator/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	app := fiber.New()
	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ethereumClient, err := eth.Dial(ctx, cfg.RPCEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	estimateService := service.NewEstimateService(logger, *ethereumClient)
	estimateHandler := handler.NewEstimateHandler(logger, estimateService)
	app.Get("/estimate", estimateHandler.Handle())

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(cfg.Addr)
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			_ = app.Shutdown()
			ethereumClient.Close()
			return fmt.Errorf("server error: %w", err)
		}
		ethereumClient.Close()
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = app.Shutdown()

	ethereumClient.Close()

	<-shutdownCtx.Done()
	return nil
}
