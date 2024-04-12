package parser

import (
	"fmt"
	"strings"
)

const (
	debugMode = !true
)

var traceLevel int = 0

const traceIdentPlaceholder string = "\t"

func identLevel() string {
	if debugMode {
		return strings.Repeat(traceIdentPlaceholder, traceLevel-1)
	}
	return ""
}

func tracePrint(fs string) {
	if debugMode {
		fmt.Printf("%s%s\n", identLevel(), fs)
	}
}

func incIdent() { traceLevel = traceLevel + 1 }
func decIdent() { traceLevel = traceLevel - 1 }

func trace(msg string) string {
	incIdent()
	tracePrint("BEGIN " + msg)
	return msg
}

func untrace(msg string) {
	tracePrint("END " + msg)
	decIdent()
}
