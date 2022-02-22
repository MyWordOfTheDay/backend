package server

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mywordoftheday/backend/internal/db"
	v1alpha1 "github.com/mywordoftheday/proto/mywordoftheday/v1alpha1"
)

func TestHeartbeat(t *testing.T) {
	s := Server{}

	t.Run("Given an initialised Server", func(t *testing.T) {
		t.Run("When a request is made to Heartbeat", func(t *testing.T) {
			t.Run("Then it should return an empty HeartbeatResponse", func(t *testing.T) {
				r, err := s.Heartbeat(context.Background(), &v1alpha1.HeartbeatRequest{})
				assert.NoError(t, err)
				assert.Equal(t, &v1alpha1.HeartbeatResponse{}, r)
			})
		})
	})
}

func TestAddWord(t *testing.T) {
	wm := &wordMock{}
	s := Server{wordModifier: wm}

	t.Run("Given a request to AddWord", func(t *testing.T) {
		t.Run("When an error is returned", func(t *testing.T) {
			t.Run("Then the error is returned to the caller", func(t *testing.T) {
				wm.err = errors.New("an error")

				r, err := s.AddWord(context.Background(), &v1alpha1.AddWordRequest{})
				assert.EqualError(t, err, "unable to add word: an error")
				assert.Nil(t, r)
			})
		})
		t.Run("When no error is returned", func(t *testing.T) {
			t.Run("Then the Word is returned to the caller", func(t *testing.T) {
				wm.err = nil
				wm.insertWordResponse = db.Word{ID: 45, Word: "floccinaucinihilipilification", CustomDefinition: "estimation of worthlessness"}

				r, err := s.AddWord(context.Background(), &v1alpha1.AddWordRequest{})
				assert.NoError(t, err)

				assert.Equal(t, wm.insertWordResponse.ID, r.Word.Id)
				assert.Equal(t, wm.insertWordResponse.Word, r.Word.Word)
				assert.Equal(t, wm.insertWordResponse.CustomDefinition, r.Word.CustomDefinition)
			})
		})
	})
}

func TestListWord(t *testing.T) {
	wm := &wordMock{}
	s := Server{wordQuerier: wm}

	t.Run("Given a request to ListWord", func(t *testing.T) {
		t.Run("When an error is returned", func(t *testing.T) {
			t.Run("Then the error is returned to the caller", func(t *testing.T) {
				wm.err = errors.New("an error")

				r, err := s.ListWords(context.Background(), &v1alpha1.ListWordsRequest{})
				assert.EqualError(t, err, "unable to list words: an error")
				assert.Nil(t, r)
			})
		})
		t.Run("When no error is returned", func(t *testing.T) {
			t.Run("Then the Words are returned to the caller", func(t *testing.T) {
				wm.err = nil
				wm.listWordsResponse = []db.Word{{ID: 45, Word: "word1"}, {ID: 1, Word: "word2", CustomDefinition: "a custom definition here"}}

				r, err := s.ListWords(context.Background(), &v1alpha1.ListWordsRequest{})
				assert.NoError(t, err)
				assert.Len(t, r.Words, 2)

				expected := make(map[int32]db.Word)
				for _, e := range wm.listWordsResponse {
					expected[e.ID] = db.Word{
						ID:               e.ID,
						Word:             e.Word,
						CustomDefinition: e.CustomDefinition,
					}
				}

				for _, a := range r.Words {
					assert.Equal(t, expected[a.Id].Word, a.Word)
					assert.Equal(t, expected[a.Id].CustomDefinition, a.CustomDefinition)
				}
			})
		})
	})
}

func TestDeleteWord(t *testing.T) {
	fm := &wordMock{}
	s := Server{wordModifier: fm}

	t.Run("Given a request to DeleteWord", func(t *testing.T) {
		t.Run("When an error is returned", func(t *testing.T) {
			t.Run("Then the error is returned to the caller", func(t *testing.T) {
				fm.err = errors.New("an error")

				r, err := s.DeleteWord(context.Background(), &v1alpha1.DeleteWordRequest{})
				assert.EqualError(t, err, "unable to delete word: an error")
				assert.Nil(t, r)
			})
		})
		t.Run("When no error is returned", func(t *testing.T) {
			t.Run("Then the Deleted Word is returned to the caller", func(t *testing.T) {
				fm.err = nil
				fm.deleteWordResponse = db.Word{ID: 45, Word: "a word", CustomDefinition: "a definition goes here"}

				r, err := s.DeleteWord(context.Background(), &v1alpha1.DeleteWordRequest{})
				assert.NoError(t, err)

				assert.Equal(t, fm.deleteWordResponse.ID, r.Word.Id)
				assert.Equal(t, fm.deleteWordResponse.Word, r.Word.Word)
				assert.Equal(t, fm.deleteWordResponse.CustomDefinition, r.Word.CustomDefinition)
			})
		})
	})
}

func TestRandomWord(t *testing.T) {
	wm := &wordMock{}
	s := Server{wordQuerier: wm}

	t.Run("Given a request to ListWords", func(t *testing.T) {
		t.Run("When an error is returned", func(t *testing.T) {
			t.Run("Then the error is returned to the caller", func(t *testing.T) {
				wm.err = errors.New("an error")

				r, err := s.RandomWord(context.Background(), &v1alpha1.RandomWordRequest{})
				assert.EqualError(t, err, "unable to get words: an error")
				assert.Nil(t, r)
			})
		})
		t.Run("When no error is returned", func(t *testing.T) {
			t.Run("Then a random word is returned", func(t *testing.T) {
				wm.err = nil
				wm.listWordsResponse = []db.Word{{ID: 45, Word: "word1"}, {ID: 1, Word: "word2", CustomDefinition: "a custom definition here"}}

				r, err := s.RandomWord(context.Background(), &v1alpha1.RandomWordRequest{})
				assert.NoError(t, err)

				isValidID := r.GetWord().GetId() == 1 || r.GetWord().GetId() == 45
				isValidWord := r.GetWord().GetWord() == "word1" || r.GetWord().GetWord() == "word2"

				assert.True(t, isValidID)
				assert.True(t, isValidWord)
			})
		})
	})
}
