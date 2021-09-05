package config

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoDBMigration(ctx context.Context, db *mongo.Database) (jc, uc *mongo.Collection) {
	var unique bool = true

	jc = db.Collection("jokes")
	uc = db.Collection("users")

	(*uc).Indexes().CreateOne(context.Background(), mongo.IndexModel{Keys: bson.M{
		"email": 1,
	}})

	(*jc).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.M{
				"id": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
		{
			Keys:    bson.M{"lang": 1},
			Options: &options.IndexOptions{},
		},
	})

	return
}

func PostGreSQLMigration(ctx context.Context, db *sqlx.DB) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}

	if _, err = tx.Exec("CREATE DATABASE IF NOT EXISTS jokeapi"); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("USE jokeapi"); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS jokes (
		id BIGSERIAL PRIMARY KEY,
		text VARCHAR(255),
		author_id BIGSERIAL DEFAULT NULL,
		explanation VARCHAR(255) DEFAULT NULL,
		lang VARCHAR(3) NOT NULL
	)`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		email VARCHAR(120),
		password VARCHAR(255),
		CONSTRAINT unique_email UNIQUE (email)
	)`); err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}

func MySQLMigration(ctx context.Context, db *sqlx.DB) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}

	if _, err = tx.Exec("CREATE DATABASE IF NOT EXISTS jokeapi"); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("USE jokeapi"); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS jokes (
		id CHAR(36) PRIMARY KEY,
		text VARCHAR(255),
		author_id VARCHAR(36) DEFAULT NULL,
		explanation VARCHAR(255) DEFAULT NULL,
		lang VARCHAR(3) NOT NULL
	)`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS joke_ratings (
		id CHAR(36) PRIMARY KEY,
		rating DECIMAL(4,2),
		joke_id VARCHAR(36)
	)`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
		id CHAR(36) PRIMARY KEY,
		email VARCHAR(120),
		password VARCHAR(255),
		CONSTRAINT unique_id UNIQUE (id)
	)`); err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}
