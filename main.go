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

	var listen bool
	flag.BoolVar(&listen, "l", false, "Listen")
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

	var conn net.Conn
	if listen {
		listener, err := net.Listen("tcp", net.JoinHostPort(host, port))

		if err != nil {
			log.Fatal(err)
		}

		conn, err = listener.Accept()

		if err != nil {
			log.Fatal(err)
		}
	} else {
		var err error
		conn, err = net.Dial("tcp", net.JoinHostPort(host, port))

		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO
	// - add udp

	remote := lines(conn)
	local := lines(os.Stdin)

	for {
		select {
		case s, ok := <-remote:
			if !ok {
				return
			}

			fmt.Print(s)
		case s := <-local:
			conn.Write([]byte(s))
		}
	}
}
