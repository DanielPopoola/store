package main

import "store/engines"

type Store struct {
	engine engines.Engine
}

func NewStore(engine engines.Engine) *Store {
	return &Store{engine: engine}
}

func (s *Store) Set(key string, value string) error {
	return s.engine.Set(key, value)
}

func (s *Store) Get(key string) (value string, err error) {
	return s.engine.Get(key)
}

func (s *Store) Delete(key string) error {
	return s.engine.Delete(key)
}
