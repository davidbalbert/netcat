package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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

	// TODO
	// - deal with control-c
	// - add listen

	remote := lines(conn)
	local := lines(os.Stdin)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	for {
		select {
		case s, ok := <-remote:
			if !ok {
				return
			}

			fmt.Print(s)
		case s := <-local:
			conn.Write([]byte(s))
		case <-sigint:
			fmt.Println()
			return
		}
	}
}
