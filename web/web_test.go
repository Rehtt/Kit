package goweb

import (
	"context"
	"fmt"
	"github.com/Rehtt/Kit/util"
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	g.SetKeyValue("test", 123)
	ctx := &Context{
		survive: true,
		Context: context.Background(),
		values:  util.DeepCopy(g.values).(map[interface{}]interface{}),
	}
	fmt.Println(ctx.Value("test"))
}
