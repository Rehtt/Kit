package util

import (
	"fmt"
	"testing"
)

func TestGetGitHubFileInfo(t *testing.T) {
	info, err := GetGitHubFileInfo("Rehtt", "Kit", "util/github.go")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", info)
}
