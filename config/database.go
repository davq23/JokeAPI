package config

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

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
		id BIGINT(10) UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		text VARCHAR(255),
		author_id BIGINT(10) UNSIGNED DEFAULT NULL,
		explanation VARCHAR(255) DEFAULT NULL,
		lang VARCHAR(3) NOT NULL
	)`); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
		id BIGINT(10) UNSIGNED PRIMARY KEY AUTO_INCREMENT,
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
