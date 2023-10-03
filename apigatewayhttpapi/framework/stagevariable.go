package framework

import (
	"fmt"
	"strings"
)

type StageVarGetter struct {
	c       *Context
	missing []string
}

func (g *StageVarGetter) Get(key string, value *string) *StageVarGetter {
	*value = g.c.StageVariable(key)
	if *value == "" {
		g.missing = append(g.missing, key)
	}

	return g
}

func (g StageVarGetter) Error() error {
	if len(g.missing) == 0 {
		return nil
	}

	return fmt.Errorf("missing %d stage variables: %s", len(g.missing), strings.Join(g.missing, ", "))
}

// Retrieves several stage variables, if any are missing then the StageVarGetter.Error() will return non-nil.
func (c *Context) StageVariables(key string, value *string) *StageVarGetter {
	getter := &StageVarGetter{c: c}
	return getter.Get(key, value)
}
