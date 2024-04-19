package getenv

import (
	"os"
)

// Variable defines ways to retrieve a string variable.
type Variable interface {
	Get() (string, error)
	MustGet() string
}

// Env calls os.Getenv and returns that value in subsequent calls.
//
// See Getenv if you need something that calls os.Getenv on every invocation.
func Env(key string) Variable {
	return fixedEnvVar{value: os.Getenv(key)}
}

// Getenv calls os.Getenv on every invocation and returns its value.
//
// Most of the time, Env suffices because environment variables are not updated that often. Use Getenv if you have a use
// case where the environment variables might be updated by some other processes.
func Getenv(key string) Variable {
	return envVar{key: key}
}

// fixedEnvVar implements Variable interface for Env.
type fixedEnvVar struct {
	value string
}

func (e fixedEnvVar) Get() (string, error) {
	return e.value, nil
}

func (e fixedEnvVar) MustGet() string {
	return e.value
}

var _ Variable = &fixedEnvVar{}
var _ Variable = (*fixedEnvVar)(nil)

// envVar implements Variable interface for Getenv.
type envVar struct {
	key string
}

func (e envVar) Get() (string, error) {
	return os.Getenv(e.key), nil
}

func (e envVar) MustGet() string {
	return os.Getenv(e.key)
}

var _ Variable = &envVar{}
var _ Variable = (*envVar)(nil)

// getter implements the Variable interface for a function.
type getter func() (string, error)

func (g getter) Get() (string, error) {
	return g()
}

func (g getter) MustGet() string {
	v, err := g()
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

func (e errVar) MustGet() string {
	panic(e.err)
}

var _ Variable = &errVar{}
var _ Variable = (*errVar)(nil)
