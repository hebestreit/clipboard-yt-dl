package main

import (
	"github.com/shivylp/clipboard"
	"golang.org/x/sys/windows"
)

func clipboardReadAll() (string, error) {
	newValue, err := clipboard.ReadAll()
	if err != nil {
		switch err {
		case windows.DS_S_SUCCESS:
			// don't throw an error if user has copied a file
			return "", nil
		default:
			return "", err
		}
	}

	return newValue, nil
}
