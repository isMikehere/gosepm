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

//root
type Root struct {
}

//reference for root
type MyData struct {
	ID   string
	name string
	*Root
}
