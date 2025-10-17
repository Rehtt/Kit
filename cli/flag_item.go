// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/10/17

package cli

import (
	"strings"
)

type FlagType uint8

const (
	FlagItemSelect FlagType = iota
	FlagItemFile
	FlagItemDir
)

type FlagItemNode struct {
	Value       string
	Description string
}

type FlagItem struct {
	Type  FlagType
	Nodes []FlagItemNode
}

func (f FlagItem) String() string {
	switch f.Type {
	case FlagItemSelect:
		keys := make([]string, 0, len(f.Nodes))
		for _, n := range f.Nodes {
			keys = append(keys, n.Value)
		}
		return "[" + strings.Join(keys, "/") + "]"
	case FlagItemFile:
		return "file"
	case FlagItemDir:
		return "dir"
	default:
		return "value"
	}
}

func NewFlagItemNode(value string, description string) FlagItemNode {
	return FlagItemNode{value, description}
}

func NewFlagItemFile() FlagItem                        { return FlagItem{FlagItemFile, nil} }
func NewFlagItemDir() FlagItem                         { return FlagItem{FlagItemDir, nil} }
func NewFlagItemSelect(nodes ...FlagItemNode) FlagItem { return FlagItem{FlagItemSelect, nodes} }
func NewFlagItemSelectString(value ...string) FlagItem {
	f := FlagItem{Type: FlagItemSelect}
	for _, v := range value {
		f.Nodes = append(f.Nodes, FlagItemNode{Value: v})
	}
	return f
}
