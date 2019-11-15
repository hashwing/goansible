package termutil

import (
	"fmt"

	"github.com/cheggaaa/pb/v3/termutil"
	"github.com/mattn/go-runewidth"
	"github.com/ttacon/chalk"
)

func Full(f, p string, a ...interface{}) string {
	s := fmt.Sprintf(f, a...)
	n := runewidth.StringWidth(s)
	w, _ := termutil.TerminalWidth()
	m := runewidth.StringWidth(p)
	c := (w - n - 4*m) / m
	for i := 0; i < c; i++ {
		s += p
	}
	return s
}

func FullPrintf(f, p string, a ...interface{}) {
	fmt.Println(Full(f, p, a...))
}

func FullInfo(f, p string, a ...interface{}) {
	fmt.Println(chalk.Blue.Color(Full(f, p, a...)))
}

func Successf(f string, a ...interface{}) {
	fmt.Println(chalk.Green.Color(fmt.Sprintf(f, a...)))
}

func Errorf(f string, a ...interface{}) {
	fmt.Println(chalk.Red.Color(fmt.Sprintf(f, a...)))
}

func Printf(f string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(f, a...))
}

func Changedf(f string, a ...interface{}) {
	fmt.Println(chalk.Yellow.Color(fmt.Sprintf(f, a...)))
}

func Infof(f string, a ...interface{}) {
	fmt.Println(chalk.Blue.Color(fmt.Sprintf(f, a...)))
}
