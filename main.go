package main

import (
	"flag"
)

var logToFile = flag.Bool("l", false, "log")

func main() {
	flag.Parse()
	if initLog(*logToFile) == false {
		return
	}
	defer closeLog()

	if initConf() == false {
		return
	}

	if initServer() == false {
		return
	}
	startServer()
	defer closeServer()
}
