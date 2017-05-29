package test

import (
	"bytes"
	"log"
)

func logg() {
	var buf bytes.Buffer

	logger := log.New(&buf, "logger: ", log.Llongfile)
	logger.Print("Hello, log file!")

	// fmt.Print(&buf)
}
