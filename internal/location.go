package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Location struct {
	File string
	Line int
}

func Loc(wd string) (Location, error) {
	var err error

	if os.Getenv("GOFILE") == "" || os.Getenv("GOLINE") == "" {
		return Location{}, fmt.Errorf("%s must be run from //go:generate", os.Args[0])
	}

	l := Location{File: filepath.Join(wd, os.Getenv("GOFILE"))}

	l.Line, err = strconv.Atoi(os.Getenv("GOLINE"))
	if err != nil {
		return Location{}, errors.Join(err, fmt.Errorf("%s must be run from //go:generate", os.Args[0]))
	}
	return l, err
}
