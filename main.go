package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davq23/jokeapi/config"
	"github.com/davq23/jokeapi/handlers"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories/mongodb"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	godotenv.Load()

	l := log.New(os.Stdout, "joke api - ", log.LstdFlags)
	cfg := config.Config{
		DBConnectionURI: os.Getenv("DB_URI"),
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.DBConnectionURI))

	if err != nil {
		l.Fatal(err.Error())
	}

	db := client.Database("jokeapi")

	jokesCollection := db.Collection("jokes")

	v := validator.New()

	vm := middlewares.NewValidation(l, v)

	jr := mongodb.NewJokeRepository(jokesCollection)

	jh := handlers.NewJoke(l, jr, vm)

	serveMux := http.NewServeMux()

	serveMux.Handle("/jokes", jh)

	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  time.Second,
		Handler:      serveMux,
		Addr:         ":8080",
	}

	go func() {
		err := server.ListenAndServe()

		if err != nil {
			l.Fatalln(err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 5)

	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan

	l.Println("Received terminate, graceful shutdown", sig)

	tc, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancelFunc()
	client.Disconnect(context.Background())

	server.Shutdown(tc)
}
