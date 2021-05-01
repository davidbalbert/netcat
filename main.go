package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// var listen = flag.Bool("l", false, "Listen")

func lines(r io.Reader) <-chan string {
	out := make(chan string)

	go func() {
		br := bufio.NewReader(r)

		for {
			s, err := br.ReadString(('\n'))

			out <- s

			if err != nil {
				close(out)
				return
			}
		}
	}()

	return out
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("usage: %s hostname port\n", os.Args[0])
	}

	var host, port string

	if flag.NArg() == 1 {
		host = ""
		port = flag.Arg(0)
	} else {
		host = flag.Arg(0)
		port = flag.Arg(1)
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))

	if err != nil {
		log.Fatal(err)
	}

	// Tomorrow
	// - make a read function that takes a reader and returns a chan string
	// - don't buffer writing
	// - select to see what's ready

	c := lines(conn)

	for s := range c {
		fmt.Print(s)
	}
}
