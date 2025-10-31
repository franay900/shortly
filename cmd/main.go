package main

import (
	"fmt"
	"net/http"
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
	server := http.Server{
		Addr:    ":8081",
		Handler: app,
	}
	fmt.Println("Server is listening on port 8081")
	server.ListenAndServe()
}
