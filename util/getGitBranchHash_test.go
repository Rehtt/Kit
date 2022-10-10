package util

import (
	"fmt"
	"testing"
)

func TestGetGitBranchHash(t *testing.T) {
	fmt.Println(GetGitBranchHash("https://github.com/Rehtt/Kit.git", "refs/heads/master"))
}
