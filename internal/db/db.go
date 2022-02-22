package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Word struct {
	ID               int32
	Word             string
	CustomDefinition string
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type Manager struct {
	pool *pgxpool.Pool
}

func New(c Config) (*Manager, error) {
	if c.Host == "" {
		return nil, errors.New("host not defined")
	}

	if c.Port == "" {
		return nil, errors.New("port not defined")
	}

	if c.Username == "" {
		return nil, errors.New("user not defined")
	}

	if c.Password == "" {
		return nil, errors.New("password not defined")
	}

	if c.Database == "" {
		return nil, errors.New("database not defined")
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.Username, c.Password, c.Host, c.Port, c.Database)

	pool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	return &Manager{pool: pool}, nil
}

func (m *Manager) Ping(ctx context.Context) error {
	return m.pool.Ping(ctx)
}

func (m *Manager) InsertWord(ctx context.Context, word Word) (Word, error) {
	w := Word{}

	err := m.pool.QueryRow(
		ctx,
		"INSERT INTO words(word, custom_definition) VALUES($1, $2) RETURNING id, word, custom_definition",
		word.Word, word.CustomDefinition,
	).Scan(&w.ID, &w.Word, &w.CustomDefinition)
	if err != nil {
		return w, errors.Wrap(err, "unable to insert word")
	}

	logrus.WithFields(logrus.Fields{
		"id": w.ID,
	}).Info("Word inserted successfully")

	return w, nil
}

func (m *Manager) ListWords(ctx context.Context) ([]Word, error) {
	words := make([]Word, 0)

	rows, err := m.pool.Query(ctx, "SELECT id, word, custom_definition FROM words")
	if err != nil {
		return words, errors.Wrap(err, "unable to get words")
	}

	rowCount := 0
	for rows.Next() {
		w := Word{}

		if err := rows.Scan(&w.ID, &w.Word, &w.CustomDefinition); err != nil {
			return nil, errors.Wrap(err, "unable to scan row")
		}

		words = append(words, w)

		rowCount++
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "erroring reading rows")
	}

	logrus.WithFields(logrus.Fields{"rowCount": rowCount}).Info("Words queried successfully")

	return words, nil
}

func (m *Manager) DeleteWord(ctx context.Context, id int32) (Word, error) {
	w := Word{}

	err := m.pool.QueryRow(
		ctx,
		"DELETE FROM words WHERE id=$1 RETURNING id, word, custom_definition",
		id,
	).Scan(&w.ID, &w.Word, &w.CustomDefinition)
	if err != nil {
		return w, errors.Wrap(err, "unable to delete word")
	}

	logrus.WithFields(logrus.Fields{
		"id": w.ID,
	}).Info("Word deleted successfully")

	return w, nil
}
