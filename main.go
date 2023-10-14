package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sabahtalateh/convert/internal"
)

func main() {
	wd, err := os.Getwd()
	check(err)

	l, err := internal.Loc(wd)
	check(err)
	checkConvertOnLine(l)

	err = internal.InitMod(wd)
	checkLoc(l, err)

	err = internal.Convert(internal.NewContext(l))
	checkLoc(l, err)
}

func checkConvertOnLine(l internal.Location) {
	bb, err := os.ReadFile(l.File)
	checkLoc(l, err)

	msg := fmt.Sprintf("no //go:generate convert\n\t%s:%d\npossible reason is multiple converts in same file. in this case rerun `go generate ..`", l.File, l.Line)

	lines := strings.Split(string(bb), "\n")
	fileLineN := l.Line - 1
	if len(lines) < fileLineN {
		fmt.Println(msg)
		os.Exit(1)
	}

	fileLine := lines[fileLineN]
	fileLineParts := strings.Fields(fileLine)

	if len(fileLineParts) < 2 {
		fmt.Println(msg)
		os.Exit(1)
	}

	if fileLineParts[0] != "//go:generate" {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func checkLoc(l internal.Location, err error) {
	if err != nil {
		fmt.Printf("convert\n%s\n\t%s:%d\n", err, l.File, l.Line)
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("convert\n%s\n", err)
		os.Exit(1)
	}
}
