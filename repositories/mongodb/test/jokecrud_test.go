package test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/davq23/jokeapi/config"
	"github.com/davq23/jokeapi/repositories/mongodb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var jokeCrud *mongodb.JokeCRUD

func TestFetchAll(t *testing.T) {
}

func TestMain(m *testing.M) {
	l := log.New(os.Stdout, "joke api - ", log.LstdFlags)

	err := godotenv.Load()

	if err != nil {
		l.Fatal(err.Error())
	}

	cfg := config.Config{
		DBConnectionURI: os.Getenv("MONGODB_URI"),
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.DBConnectionURI))

	if err != nil {
		l.Fatal(err.Error())
	}

	db := client.Database("jokeapi")

	_, _ = config.MongoDBMigration(context.Background(), db)

	exitVal := m.Run()

	client.Disconnect(context.Background())

	os.Exit(exitVal)
}
