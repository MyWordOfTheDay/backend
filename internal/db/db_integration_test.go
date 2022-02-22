//go:build integration
// +build integration

package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/mywordoftheday/backend/internal/db"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

var mgr *db.Manager

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	fmt.Println("Creating test container...")

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=username",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://username:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	var conn *sql.DB

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		conn, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return conn.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := createTables(conn); err != nil {
		log.Fatalf("Could not create tables: %s", err)
	}

	mgr, err = db.New(db.Config{
		Host:     "localhost",
		Port:     resource.GetPort("5432/tcp"),
		Username: "username",
		Password: "secret",
		Database: "dbname",
	})
	if err != nil {
		log.Fatalf("Could not create new db instance: %s", err)
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(conn *sql.DB) error {
	// Words Table
	query := `CREATE TABLE IF NOT EXISTS "words" (
  "id" SERIAL PRIMARY KEY NOT NULL,
  "word" VARCHAR(255) DEFAULT '',
  "custom_definition" VARCHAR(255) DEFAULT ''
	);`

	if _, err := conn.Exec(query); err != nil {
		return err
	}

	return nil
}

func TestPing(t *testing.T) {
	t.Run("Given a initialised db manager", func(t *testing.T) {
		t.Run("When ping is called", func(t *testing.T) {
			t.Run("Then no error is returned", func(t *testing.T) {
				assert.NoError(t, mgr.Ping(context.Background()))
			})
		})
	})
}

func TestWords(t *testing.T) {
	t.Run("Given a valid Word object", func(t *testing.T) {
		var inserted db.Word
		var err error

		word := db.Word{Word: "floccinaucinihilipilification"}

		t.Run("When it is passed to InsertWord", func(t *testing.T) {
			t.Run("Then it should create the record without error", func(t *testing.T) {
				inserted, err = mgr.InsertWord(context.Background(), word)
				assert.NoError(t, err)
				assert.NotNil(t, inserted)

				assert.Equal(t, word.Word, inserted.Word)
				assert.Equal(t, word.CustomDefinition, inserted.CustomDefinition)
			})
		})

		t.Run("When ListWord is called", func(t *testing.T) {
			t.Run("Then the inserted Word should exist", func(t *testing.T) {
				w, err := mgr.ListWords(context.Background())
				assert.NoError(t, err)

				assert.Len(t, w, 1)

				assert.Equal(t, inserted.ID, w[0].ID)
				assert.Equal(t, word.Word, w[0].Word)
				assert.Equal(t, word.CustomDefinition, w[0].CustomDefinition)
			})
		})

		t.Run("When DeleteWord is called", func(t *testing.T) {
			t.Run("Then the Word is deleted", func(t *testing.T) {
				f, err := mgr.DeleteWord(context.Background(), inserted.ID)
				assert.NoError(t, err)
				assert.NotNil(t, f)

				assert.Equal(t, inserted.ID, f.ID)
				assert.Equal(t, word.Word, f.Word)
				assert.Equal(t, word.CustomDefinition, f.CustomDefinition)

				// Make sure the word doesn't exist
				lf, err := mgr.ListWords(context.Background())
				assert.NoError(t, err)

				assert.Len(t, lf, 0)
			})
		})
	})
}
