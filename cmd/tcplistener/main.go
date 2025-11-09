package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func(ch chan<- string, f io.ReadCloser) {
		defer close(ch)
		defer f.Close()
		s := ""
		for {
			sl := make([]byte, 8, 8)
			n, _ := f.Read(sl)
			if n == 0 {
				break
			}
			s += string(sl)
			broken := strings.Split(s, "\n")
			if len(broken) == 1 {
				s = broken[0]
			} else {
				ch <- broken[0]
				s = broken[1]
			}
		}
		if len(s) > 0 {
			ch <- s
		}
	}(lines, f)
	return lines
}

func main() {
	listner, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal("error", err.Error())
		return
	}
	defer listner.Close()
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal("error", err.Error())
		}
		log.Println("Connection accepted successfully")

		ch := getLinesChannel(conn)
		for s := range ch {
			fmt.Println(s)
		}
		log.Println("Connection closed")

	}
}
