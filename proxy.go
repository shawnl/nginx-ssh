package main

import (
	"os"
	"net"
	"io"
	"fmt"
	"bufio"
	"bytes"
)

func copyAndClose(w io.Writer, r io.ReadCloser) {
	io.Copy(w, r)
	r.Close()
}

func handleConnection(c net.Conn) {
	r := bufio.NewReader(c)
	buf, err := r.Peek(4)
	var d net.Conn
	if bytes.Equal(buf, []byte{'S', 'S', 'H', '-'}) {
		d, err = net.Dial("tcp", "localhost:22")
	} else if bytes.Equal(buf[:2], []byte{0x16, 0x03}) && buf[3] >= 0x00 && buf[3] <= 0x03 {
		d, err = net.Dial("tcp", "localhost:443")
	}
	if err != nil {
		fmt.Println(err)
		c.Close()
		return
	}
	go copyAndClose(c, d)
	io.Copy(d, r)
	c.Close()
}

func main() {
	ln, err := net.Listen("tcp", ":4443")
	if err != nil {
		fmt.Println("Listen failed: ", err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept failed: ", err)
			continue
		}
		go handleConnection(conn)
	}
}
