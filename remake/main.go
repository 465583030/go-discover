package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

var wg sync.WaitGroup

func RequestHTTP(ip string, port string) (bool, string) {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	return true, buf.String()
}

func main() {
	a, b := RequestHTTP("google.com", "80")
	fmt.Println(a)
	fmt.Println(b)
}
