package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

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
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	addr, err := func() (net.Addr, error) {
		for _, iface := range ifaces {
			addrs, ifaceErr := iface.Addrs()
			if ifaceErr != nil {
				log.Fatal(err)
			}
			for _, addr := range addrs {
				oip, _, cidrErr := net.ParseCIDR(addr.String())
				if cidrErr != nil {
					log.Fatal(err)
				}
				if str := strings.HasPrefix(addr.String(), "127"); str == false {
					if oip.To4() != nil {
						return addr, nil
					}
				}
			}
		}
		return nil, errors.New("Error occured in getting CIDR.")
	}()

	if err != nil {
		panic(err)
	}

	oip, ipnet, err := net.ParseCIDR(addr.String())

	for ip := oip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		a, b := RequestHTTP(ip.String(), "")
		if a {
			fmt.Println(b)
		}
	}
}
