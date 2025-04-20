package slice

import (
	"errors"

	"github.com/Rehtt/Kit/strings"
)

type FlagSetArray[T any] struct {
	array []T

	HandleSet func(v string) T
}

func (f FlagSetArray[T]) String() string {
	return strings.JoinToString(f.array, ",")
}

func (f *FlagSetArray[T]) Set(v string) error {
	if f.HandleSet != nil {
		f.array = append(f.array, f.HandleSet(v))
		return nil
	} else {
		return errors.New("HandleSet is nil")
	}
}

func (f *FlagSetArray[T]) Get() any {
	return f.array
}
