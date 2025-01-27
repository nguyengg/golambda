package getenv

import (
	"context"
	"os"
)

// Variable defines ways to retrieve a variable.
type Variable[T any] interface {
	Get() (T, error)
	GetWithContext(ctx context.Context) (T, error)
	MustGet() T
	MustGetWithContext(ctx context.Context) T
}

// Env calls os.Getenv and returns that value in subsequent calls.
//
// See Getenv if you need something that calls os.Getenv on every invocation.
func Env(key string) Variable[string] {
	v := os.Getenv(key)
	return Getter(func(ctx context.Context) (string, error) {
		return v, nil
	})
}

// Getenv calls os.Getenv on every invocation and returns its value.
//
// Most of the time, Env suffices because environment variables are not updated that often. Use Getenv if you have a use
// case where the environment variables might be updated by some other processes.
func Getenv(key string) Variable[string] {
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
