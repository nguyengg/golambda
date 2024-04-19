package getenv

import (
	"context"
	"os"
)

// Variable defines ways to retrieve a string variable.
type Variable interface {
	Get() (string, error)
	GetWithContext(ctx context.Context) (string, error)
	MustGet() string
	MustGetWithContext(ctx context.Context) string
}

// Env calls os.Getenv and returns that value in subsequent calls.
//
// See Getenv if you need something that calls os.Getenv on every invocation.
func Env(key string) Variable {
	v := os.Getenv(key)
	return Getter(func(ctx context.Context) (string, error) {
		return v, nil
	})
}

// Getenv calls os.Getenv on every invocation and returns its value.
//
// Most of the time, Env suffices because environment variables are not updated that often. Use Getenv if you have a use
// case where the environment variables might be updated by some other processes.
func Getenv(key string) Variable {
	return Getter(func(ctx context.Context) (string, error) {
		return os.Getenv(key), nil
	})
}

// Getter implements the Variable interface for a function.
type Getter func(ctx context.Context) (string, error)

func (g Getter) Get() (string, error) {
	return g.GetWithContext(context.Background())
}

func (g Getter) GetWithContext(ctx context.Context) (string, error) {
	return g(ctx)
}

func (g Getter) MustGet() string {
	return g.MustGetWithContext(context.Background())
}

func (g Getter) MustGetWithContext(ctx context.Context) string {
	v, err := g(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// err implements the Variable interface for an error.
type errVar struct {
	err error
}

func (e errVar) Get() (string, error) {
	return "", e.err
}

func (e errVar) GetWithContext(_ context.Context) (string, error) {
	return "", e.err
}

func (e errVar) MustGet() string {
	panic(e.err)
}

func (e errVar) MustGetWithContext(_ context.Context) string {
	panic(e.err)
}

var _ Variable = &errVar{}
var _ Variable = (*errVar)(nil)
