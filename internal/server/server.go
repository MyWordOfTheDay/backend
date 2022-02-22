package server

import (
	"context"

	"github.com/mywordoftheday/backend/internal/db"
	v1alpha1 "github.com/mywordoftheday/proto/mywordoftheday/v1alpha1"
	"github.com/pkg/errors"
)

type wordQuerier interface {
	ListWords(context.Context) ([]db.Word, error)
}

type wordModifier interface {
	InsertWord(context.Context, db.Word) (db.Word, error)
	DeleteWord(context.Context, int32) (db.Word, error)
}

// Server is the implementation of the mywordofthedayv1alpha1.MyWordOfTheDayServer
type Server struct {
	wordQuerier  wordQuerier
	wordModifier wordModifier
}

type Config struct {
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBName     string
}

func New(c Config) (*Server, error) {
	dbManager, err := db.New(db.Config{
		Host:     c.DBHost,
		Port:     c.DBPort,
		Username: c.DBUsername,
		Password: c.DBPassword,
		Database: c.DBName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create db instance")
	}

	return &Server{
		wordQuerier:  dbManager,
		wordModifier: dbManager,
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *v1alpha1.HeartbeatRequest) (*v1alpha1.HeartbeatResponse, error) {
	return &v1alpha1.HeartbeatResponse{}, nil
}

func (s *Server) AddWord(ctx context.Context, req *v1alpha1.AddWordRequest) (*v1alpha1.AddWordResponse, error) {
	rsp, err := s.wordModifier.InsertWord(ctx, db.Word{
		Word:             req.GetWord().GetWord(),
		CustomDefinition: req.GetWord().GetCustomDefinition(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to add word")
	}

	return &v1alpha1.AddWordResponse{
		Word: &v1alpha1.Word{
			Id:               rsp.ID,
			Word:             rsp.Word,
			CustomDefinition: rsp.CustomDefinition,
		},
	}, nil
}

func (s *Server) ListWords(ctx context.Context, req *v1alpha1.ListWordsRequest) (*v1alpha1.ListWordsResponse, error) {
	rsp, err := s.wordQuerier.ListWords(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list words")
	}

	w := make([]*v1alpha1.Word, len(rsp))
	for i, word := range rsp {
		w[i] = &v1alpha1.Word{
			Id:               word.ID,
			Word:             word.Word,
			CustomDefinition: word.CustomDefinition,
		}
	}

	return &v1alpha1.ListWordsResponse{
		Words: w,
	}, nil
}

func (s *Server) DeleteWord(ctx context.Context, req *v1alpha1.DeleteWordRequest) (*v1alpha1.DeleteWordResponse, error) {
	rsp, err := s.wordModifier.DeleteWord(ctx, req.GetId())
	if err != nil {
		return nil, errors.Wrap(err, "unable to delete word")
	}

	return &v1alpha1.DeleteWordResponse{
		Word: &v1alpha1.Word{
			Id:               rsp.ID,
			Word:             rsp.Word,
			CustomDefinition: rsp.CustomDefinition,
		},
	}, nil
}
