//go:build !windows

package util

import (
	"log"
	"os"
)

func DeleteSelf() {
	if err := os.Remove(os.Args[0]); err != nil {
		log.Printf("Delete Self Error: %d\n", err)
	}
}
