// +build darwin dragonfly freebsd linux netbsd openbsd

package main

import (
	"github.com/shivylp/clipboard"
)

func clipboardReadAll() (string, error) {
	return clipboard.ReadAll()
}
