package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url/short/configs"
	"url/short/internal/auth"
	"url/short/internal/link"
	"url/short/internal/stat"
	"url/short/internal/user"
	"url/short/pkg/db"
	"url/short/pkg/event"
	"url/short/pkg/middleware"
)

func App() http.Handler {
	conf := configs.LoadConfig()
	DB := db.NewDB(conf)
	router := http.NewServeMux()
	eventBus := event.NewEventBus()

	// Repositories
	linkRepository := link.NewLinkRepository(DB)
	userRepository := user.NewUserRepository(DB)
	statRepository := stat.NewStatRepository(DB)

	// Services
	authService := auth.NewAuthService(userRepository)
	statService := stat.NewStatService(&stat.StatServiceDeps{
		EventBus:       eventBus,
		StatRepository: statRepository,
	})

	// Handler
	auth.NewAuthHandler(router, auth.AuthHandlerDeps{
		Config:      conf,
		AuthService: authService,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		LinkRepository: linkRepository,
		Config:         conf,
		EventBus:       eventBus,
	})
	stat.NewStatHandler(router, stat.StatHandlerDeps{
		StatRepository: statRepository,
		Config:         conf,
	})

	go statService.AddClick()

	// Middlewares
	stack := middleware.Chain(
		middleware.Cors,
		middleware.Logging,
	)
	return stack(router)
}

func main() {
	app := App()
	server := &http.Server{
		Addr:    ":8081",
		Handler: app,
	}

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Server is listening on port 8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
