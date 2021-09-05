package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
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

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	godotenv.Load()

	l := log.New(os.Stdout, "joke api - ", log.LstdFlags)
	cfg := config.Config{
		DBConnectionURI: os.Getenv("MONGODB_URI"),
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.DBConnectionURI))

	if err != nil {
		l.Fatal(err.Error())
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		l.Fatal(err.Error())
	}

	db := client.Database("jokeapi")

	jc, uc := config.MongoDBMigration(context.Background(), db)

	jr := mongodb.NewJoke(jc)
	ur := mongodb.NewUser(uc)

	if err != nil {
		l.Fatal(err.Error())
	}

	v := validator.New()

	idRegexp := regexp.MustCompile(`[a-fA-F\d]{24}`)

	vm := middlewares.NewValidation(l, v)

	v.RegisterValidation("joke_id", jr.CheckValidID)
	v.RegisterValidation("user_id", ur.CheckValidID)
	v.RegisterValidation("rating_id", ur.CheckValidID)
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		return idRegexp.Match([]byte(fl.Field().String()))
	})

	am := middlewares.NewAuth(l, []byte(os.Getenv("API_KEY")))

	jh := handlers.NewJoke(l, jr, vm, am)
	jrh := handlers.NewJokeRating(l, jr, vm, am)
	uh := handlers.NewUser(l, ur, vm, am)
	ah := handlers.NewAuth(am, l, ur, vm)

	serveMux := http.NewServeMux()

	serveMux.Handle("/jokes/ratings", jrh)
	serveMux.Handle("/jokes/ratings/", jrh)
	serveMux.Handle("/jokes", jh)
	serveMux.Handle("/jokes/", jh)
	serveMux.Handle("/users", uh)
	serveMux.Handle("/users/", uh)
	serveMux.Handle("/login", ah)

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
