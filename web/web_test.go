package goweb

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	g.SetKeyValue("test", 123)
	g.Grep("/123")
	g.Grep("/345")
}
