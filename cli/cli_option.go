// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/20

package cli

import (
	"fmt"
	"slices"
)

type commandSort uint8

const (
	// CommandSortAdded 表示按照添加顺序排序命令
	CommandSortAdded commandSort = iota
	// CommandSortAlphaAsc 表示按照字母升序排序命令
	CommandSortAlphaAsc
	// CommandSortAlphaDesc 表示按照字母降序排序命令
	CommandSortAlphaDesc
)

type SubCommands struct {
	commands    []*CLI
	commandsMap map[string]*CLI
	sort        commandSort
}

func (s *SubCommands) SetSort(sort commandSort) {
	s.sort = sort
}

func (s *SubCommands) GetSort() commandSort {
	return s.sort
}

func (s *SubCommands) Len() int {
	return len(s.commands)
}

func (s *SubCommands) Get(use string) *CLI {
	if s.commandsMap == nil {
		return nil
	}
	return s.commandsMap[use]
}

func (s *SubCommands) Add(cli ...*CLI) error {
	if s.commandsMap == nil {
		s.commandsMap = make(map[string]*CLI, len(cli))
	}
	if cap(s.commands) == 0 {
		s.commands = make([]*CLI, 0, len(cli))
	}
	for _, v := range cli {
		if _, ok := s.commandsMap[v.Use]; ok {
			return fmt.Errorf("duplicate command: %s", v.Use)
		}
		s.commandsMap[v.Use] = v
		s.commands = append(s.commands, v)
	}
	return nil
}

func (s *SubCommands) CloneList() []*CLI {
	return slices.Clone(s.commands)
}

func (s *SubCommands) Range(f func(cli *CLI) bool) {
	for _, c := range s.commands {
		if !f(c) {
			return
		}
	}
}
