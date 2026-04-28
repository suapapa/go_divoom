package main

import (
	"flag"
	"log"
)

func main() {
	var mode string
	var addr string

	flag.StringVar(&mode, "mode", "stdio", "transport mode: stdio or sse")
	flag.StringVar(&addr, "addr", "127.0.0.1:8080", "listen address for sse mode")
	flag.Parse()

	srv := newMCPServer()

	switch mode {
	case "stdio":
		if err := serveStdio(srv); err != nil {
			log.Fatalf("stdio server failed: %v", err)
		}
	case "sse":
		if err := serveSSE(srv, addr); err != nil {
			log.Fatalf("sse server failed: %v", err)
		}
	default:
		log.Fatalf("invalid mode %q (expected stdio or sse)", mode)
	}
}
