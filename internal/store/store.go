package store

import (
	"errors"
	"sync"

	"featureflag-api/internal/model"
)

var ErrKeyExists = errors.New("flag key already exists")

type Store struct {
	mu    sync.RWMutex
	flags map[string]model.Flag
}

func New() *Store {
	return &Store{
		flags: make(map[string]model.Flag),
	}
}

func (s *Store) Create(f model.Flag) error {
	return errors.New("not implemented")
}

func (s *Store) List() []model.Flag {
	return []model.Flag{}
}

func (s *Store) Get(key string) (model.Flag, bool) {
	return model.Flag{}, false
}

func (s *Store) Update(key string, f model.Flag) bool {
	return false
}

func (s *Store) Delete(key string) bool {
	return false
}
