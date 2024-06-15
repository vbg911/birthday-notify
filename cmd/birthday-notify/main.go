package main

import (
	"birthday-notify/internal/app/handlers"
	"birthday-notify/internal/app/middleware"
	"birthday-notify/internal/config"
	"birthday-notify/internal/services"
	"birthday-notify/internal/sl"
	"birthday-notify/internal/storage"
	"birthday-notify/internal/storage/userRepo"
	"github.com/gofiber/fiber/v2"
	middlewareLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/lmittmann/tint"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//init config
	cfg := config.MustReadConfig()

	//init logger
	logger := sl.SetupLogger(cfg.Env)
	logger.Info("[app] Starting birthday-notify", slog.String("env", cfg.Env))

	//init storage
	db, err := storage.NewPoolWithPing(cfg.PostgresConnect)
	if err != nil {
		logger.Error("[app] Failed to init storage", tint.Err(err))
		os.Exit(1)
	}
	logger.Info("[app] Storage init successfully")

	//init user repo
	userRepository := userRepo.NewUserRepo(db)
	logger.Info("[app] User repository created successfully")

	app := fiber.New(fiber.Config{
		AppName:               "birthday-notify",
		DisableStartupMessage: true,
		ReadTimeout:           cfg.ReadTimeout,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middlewareLogger.New(middlewareLogger.Config{
		Format:     "${black}${time}${reset} ${blue}RequestID: ${locals:requestid}${reset}  ${magenta}ExecutedTime: ${latency}${reset} Status: ${status} - ${method} ${path}\n",
		TimeFormat: time.StampMilli,
	}))

	userHandler := handlers.NewUserHandler(userRepository)
	app.Post("/register", userHandler.CreateUser)
	app.Post("/login", userHandler.Login)
	app.Post("/subscribe", middleware.Protected(), userHandler.Subscribe)
	app.Post("/unsubscribe", middleware.Protected(), userHandler.UnSubscribe)
	if cfg.Env != "prod" {
		app.Get("/metrics", monitor.New(monitor.Config{Title: "birthday-notify Metrics Page"}))
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	//init cron
	c := cron.New()
	_, err = c.AddFunc("@daily", func() { services.RunNotifier(userRepository, cfg.SMTPServer, logger) })
	if err != nil {
		logger.Error("[app] Failed to init cron", tint.Err(err))
		os.Exit(1)
	}
	c.Start()

	// запуск рассылки при первом запуске, дальше автоматически
	go func() {
		services.RunNotifier(userRepository, cfg.SMTPServer, logger)
	}()

	//init fiber server
	go func() {
		err = app.Listen(cfg.FiberAddress)
		if err != nil {
			logger.Error("[app] Failed to start fiber", tint.Err(err))
			os.Exit(1)
		}
	}()
	logger.Info("[app] Fiber started successfully", slog.String("address", cfg.FiberAddress))
	_ = <-done
	logger.Info("[app] Gracefully shutting down birthday-notify...")
	err = app.Shutdown()
	if err != nil {
		logger.Warn("[app] Failed to shutting down fiber gracefully", slog.String("error", err.Error()))
	}
	db.Close()
	c.Stop()
	logger.Info("[app] birthday-notify stopped")
}
