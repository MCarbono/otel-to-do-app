package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"otel-to-do-app/config"
	"otel-to-do-app/mongodb"
	"otel-to-do-app/router"
	"otel-to-do-app/tracing"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var (
	env = flag.String("env", "local", "used to know what environment the project is running")
)

func main() {
	flag.Parse()
	cfg, err := config.LoadEnvConfig(*env)
	if err != nil {
		panic(err)
	}
	apiServer := http.Server{Addr: fmt.Sprintf(":%v", cfg.ApiServerPort)}
	todoApiServer := http.Server{Addr: fmt.Sprintf(":%v", cfg.TodoServerPort)}
	ch := make(chan error)
	defer close(ch)
	go newAPIServer(ch, apiServer, cfg.TracingExporterURL)
	err = <-ch
	if err != nil {
		panic(err)
	}
	go newTodoAPIServer(ch, todoApiServer, cfg.TracingExporterURL, cfg.MongoDBURI)
	err = <-ch
	if err != nil {
		shutDownCtx, shutdowRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdowRelease()
		if err := apiServer.Shutdown(shutDownCtx); err != nil {
			fmt.Printf("apiServer shutdown error: %v", err.Error())
		}
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutDownCtx, shutdowRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdowRelease()
	if err := apiServer.Shutdown(shutDownCtx); err != nil {
		fmt.Printf("apiServer shutdown error: %v", err.Error())
	}
	fmt.Println("apiServer terminated successfully")
	if err := todoApiServer.Shutdown(shutDownCtx); err != nil {
		fmt.Printf("todoApiServer shutdown error: %v", err.Error())
	}
	fmt.Println("todoApiServer terminated successfully")
}

func newAPIServer(ch chan<- error, apiServer http.Server, tracingExporterURL string) {
	tp, err := tracing.JaegerTracingProvider("api-service", tracingExporterURL)
	if err != nil {
		ch <- err
		return
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	ch <- nil
	apiServer.Handler = router.NewRouterAPI()
	if err := apiServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("stopped serving connection for apiServer")
			return
		}
		fmt.Printf("apiServer shutdown error: %v", err.Error())
	}
}

func newTodoAPIServer(ch chan<- error, server http.Server, tracingExporterURL, mongoDBURI string) {
	tp, err := tracing.JaegerTracingProvider("todo-service", tracingExporterURL)
	if err != nil {
		ch <- err
		return
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	client, err := mongodb.NewMongoDB(mongoDBURI)
	if err != nil {
		ch <- err
		return
	}
	ch <- nil
	server.Handler = router.NewRouterTodoAPP(client)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("stopped serving new connections for todoApiServer.")
			return
		}
		log.Fatal("todoApiServer shutdown error: %w", err)
	}
}
