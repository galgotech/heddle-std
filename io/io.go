package io

import (
	"fmt"
	"os"
)

func Print(value string) error {
	_, err := fmt.Fprintln(os.Stdout, value)
	return err
}

func PrintError(err error) error {
	_, werr := fmt.Fprintln(os.Stderr, err)
	return werr
}
