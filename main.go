package main

// func main() {
// 	srv1 := &http.Server{
// 		Addr:    ":3000",
// 		Handler: newRouter(),
// 	}

// 	srv2 := &http.Server{
// 		Addr:    ":3001",
// 		Handler: newRouter(),
// 	}

// 	go func() {
// 		if err := srv1.ListenAndServe(); err != nil {
// 			if errors.Is(err, http.ErrServerClosed) {
// 				log.Println("stopped serving new connections for srv1.")
// 				shutDownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
// 				defer shutdownRelease()
// 				if err := srv2.Shutdown(shutDownCtx); err != nil {
// 					log.Fatalf("HTTP shutdown error for srv2: %v", err)
// 				}
// 				log.Println("Graceful shutdown completed for srv2.")
// 			}
// 		}
// 	}()

// 	go func() {
// 		if err := srv2.ListenAndServe(); err != nil {
// 			if errors.Is(err, http.ErrServerClosed) {
// 				log.Println("stopped serving new connections for srv2.")
// 			}
// 		}
// 	}()

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	<-sigChan

// 	shutDownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer shutdownRelease()
// 	if err := srv1.Shutdown(shutDownCtx); err != nil {
// 		log.Fatalf("HTTP shutdown error: %v", err)
// 	}
// 	log.Println("Graceful shutdown complete.")
// }

// func newRouter() *gin.Engine {
// 	r := gin.Default()
// 	r.GET("/", func(ctx *gin.Context) {
// 		ctx.JSON(http.StatusOK, "Hello world")
// 	})
// 	return r
// }

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"otel-to-do-app/mongodb"
	"otel-to-do-app/router"
	"otel-to-do-app/tracing"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	apiServer := http.Server{Addr: ":8888"}
	todoApiServer := http.Server{Addr: ":8080"}

	go func() {
		tp, err := tracing.JaegerTracingProvider("todo-service", "http://localhost:14268/api/traces")
		if err != nil {
			panic(err)
		}
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		client, err := mongodb.NewMongoDB("mongodb://localhost:27017")
		if err != nil {
			panic(err)
		}
		todoApiServer.Handler = router.NewRouterTodoAPP(client)
		if err := todoApiServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Println("stopped serving new connections for todoApiServer.")
				return
			}
			log.Fatal("todoApiServer shutdown error: %w", err)
		}
	}()

	time.Sleep(2 * time.Second)

	go func() {
		tp, err := tracing.JaegerTracingProvider("api-service", "http://localhost:14268/api/traces")
		if err != nil {
			panic(err)
		}
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		apiServer.Handler = router.NewRouterAPI()
		if err := apiServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("stopped serving connection for apiServer. Started to shutting down connections to todoApiServer")
			} else {
				fmt.Printf("apiServer shutdown error: %v", err.Error())
				fmt.Println("Started to shutting down connections to todoApiServer")
			}
			shutDownCtx, shutdowRelease := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdowRelease()
			if err := todoApiServer.Shutdown(shutDownCtx); err != nil {
				fmt.Printf("todoApiServer shutdown error: %v", err.Error())
				return
			}
			fmt.Println("graceful shutdow completed for todoApiServer")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutDownCtx, shutdowRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdowRelease()
	if err := apiServer.Shutdown(shutDownCtx); err != nil {
		log.Fatal("apiServer shutdown error: %w", err)
	}
	log.Println("graceful shutdown completed for apiServer")
}
