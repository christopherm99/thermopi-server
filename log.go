package main

import (
	"fmt"
	"log"
	"os"
)

func color(s string, c int) string {
	var code int
	switch c {
	case -1:
		code = 31
	case 0:
		code = 33
	case 1:
		code = 34
	case 2:
		code = 36
	case 3:
		code = 35
	default:
		code = 37
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, s) // Changes the color.
}

// -1: Fatal Error. 0: Non-fatal Error. 1: Normal Message. 2: Debug Message. 3: Verbose Message.
func logf(level int, format string, a ...interface{}) {
	var message string
	if level == -1 {
		log.Fatalln(color("(EE) "+fmt.Sprintf(format, a), level))
	} else if level <= config.verbosity {
		switch level {
		case 0:
			message = "(WW) " + fmt.Sprintf(format, a)
		case 1:
			message = "(II) " + fmt.Sprintf(format, a)
		case 2:
			message = "(DD) " + fmt.Sprintf(format, a)
		case 3:
			message = "(VV) " + fmt.Sprintf(format, a)
		default:
			logf(1, format, a)
		}
	}
	log.Println(color(message, level))
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logf(0, "%s", err)
	}
	if _, err := f.Write([]byte(message)); err != nil {
		logf(0, "%s", err)
	}
	if err := f.Close(); err != nil {
		logf(0, "%s", err)
	}
}
