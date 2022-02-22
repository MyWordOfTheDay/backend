package server

import (
	"context"

	"github.com/mywordoftheday/backend/internal/db"
)

type wordMock struct {
	insertWordResponse db.Word
	deleteWordResponse db.Word
	listWordsResponse  []db.Word
	err                error
}

func (f wordMock) InsertWord(context.Context, db.Word) (db.Word, error) {
	return f.insertWordResponse, f.err
}

func (f wordMock) DeleteWord(context.Context, int32) (db.Word, error) {
	return f.deleteWordResponse, f.err
}

func (f wordMock) ListWords(context.Context) ([]db.Word, error) {
	return f.listWordsResponse, f.err
}
