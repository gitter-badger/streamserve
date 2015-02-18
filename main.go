package main

import (
	"errors"
	"flag"
	"log"
	"runtime"
	"time"
)

type Config struct {
	Addr            string
	Path            string
	FrameBytes      uint64
	HeaderBytes     uint64
	SourceBuffer    uint64
	StatLogInterval time.Duration
	CloseIdle       bool
}

var config Config

func init() {
	config = Config{}
	c := &config
	flag.StringVar(&c.Addr, "address", "0.0.0.0:80",
		"Address to listen on: \"host:port\" where host and port can be names or numbers.")
	flag.StringVar(&c.Path, "path", "/dev/stdin",
		"Path to a source fifo, or a directory containing source fifos mapped onto the URI namespace.")
	flag.Uint64Var(&c.FrameBytes, "frame-bytes", 64,
		"Size of a data frame. Only complete frames are sent to clients.")
	flag.Uint64Var(&c.HeaderBytes, "header-bytes", 0,
		"Size of header. A header is read from each source when it is opened, and delivered to each client before sending any data bytes.")
	flag.BoolVar(&c.CloseIdle, "close-idle", false,
		"Close an input FIFO if all of its clients disconnect. This stops whatever process is writing to the FIFO, which can be useful if that process consumes resources, but depends on that process to restart/resume reliably. The FIFO will reopen next time a client requests it.")
	flag.Uint64Var(&c.SourceBuffer, "source-buffer", 64,
		"Number of frames to keep in memory for each source. The smaller this buffer is, the sooner a slow client will miss frames.")
	flag.DurationVar(&c.StatLogInterval, "stat-log-interval", 0,
		"Seconds between periodic statistics logs for each stream source, or 0 to disable.")
}

func (c Config) Check() error {
	if c.SourceBuffer <= 2 {
		return errors.New("-source-buffer must be greater than 2.")
	}
	if c.FrameBytes < 1 {
		return errors.New("-frame-bytes must not be zero.")
	}
	if c.Path == "" {
		return errors.New("-path must not be empty.")
	}
	return nil
}

func main() {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))
	flag.Parse()
	if err := config.Check(); err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}
	listening := make(chan string)
	go func() { log.Printf("Listening at %s", <-listening) }()
	log.Fatal(RunNewServer(config, listening, nil))
}
