package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/appthrust/kutelog/pkg/core"
	"github.com/appthrust/kutelog/pkg/emitters/fanout"
	"github.com/appthrust/kutelog/pkg/emitters/stdout"
	"github.com/appthrust/kutelog/pkg/emitters/websocket"
	"github.com/appthrust/kutelog/pkg/parsers/logr"
	"github.com/appthrust/kutelog/pkg/parsers/multiple"
	"github.com/appthrust/kutelog/pkg/receriver"
	"github.com/appthrust/kutelog/pkg/version"
)

func main() {
	showVersion := flag.Bool("version", false, "show version")
	verbose := flag.Bool("verbose", false, "enable verbose output")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", version.Name, version.Version)
		os.Exit(0)
	}

	// Initialize parsers
	logrParser := logr.NewParser()
	multiParser := multiple.NewParser(logrParser)

	// Initialize receiver with multi-parser
	receiver := receriver.NewReceiver(multiParser)

	// Initialize emitters
	wsEmitter := websocket.NewEmitter()
	var emitter core.Emitter
	if *verbose {
		stdoutEmitter := stdout.NewEmitter()
		emitter = fanout.NewEmitter(wsEmitter, stdoutEmitter)
	} else {
		emitter = fanout.NewEmitter(wsEmitter)
	}

	// Create and start process
	process := core.NewProcess(&core.ProcessOptions{
		Receiver: receiver,
		Emitter:  emitter,
	})

	if err := process.Start(); err != nil {
		log.Fatal(err)
	}
}
