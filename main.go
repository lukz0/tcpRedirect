package main

import (
	"net"
	"os"
	"sync"
)

// This program uses os.Stdout instead of fmt.Println because including fmt increases binary size
// Non stripped and non upx compressed sizes:
// Without fmt: 2169344
// With fmt: 2588672

func main() {
	if len(os.Args) != 3 {
		os.Stdout.Write([]byte("Usage: tcpREdirect source target\nsource and target = :port, ipv4:port or [ipv6]:port"))
		return
	}
	sourceAddr, err := net.ResolveTCPAddr("tcp", os.Args[1])
	if err != nil {
		os.Stdout.Write(append(append([]byte("Error parsing source address: "), err.Error()...), '\n'))
	}
	targetAddr, err := net.ResolveTCPAddr("tcp", os.Args[2])
	if err != nil {
		os.Stdout.Write(append(append([]byte("Error parsing target address: "), err.Error()...), '\n'))
	}

	listener, err := net.ListenTCP("tcp", sourceAddr)
	if err != nil {
		panic(err)
	}
	//fmt.Println("Proxying from", listener.Addr(), "to", targetAddr)
	os.Stdout.Write(append(append(append(append([]byte("Proxying from "), listener.Addr().String()...), " to "...), targetAddr.String()...), '\n'))
	for {
		sourceConn, sourceErr := listener.AcceptTCP()
		targetConn, targetErr := net.DialTCP("tcp", nil, targetAddr)
		if sourceErr != nil || targetErr != nil {
			// Sould the program do something?
		} else {
			go handleConn(sourceConn, targetConn)
		}
	}
}

func handleConn(conn1 *net.TCPConn, conn2 *net.TCPConn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		for {
			_, err := conn1.ReadFrom(conn2)
			if err != nil {
				break
			}
		}
		wg.Done()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		for {
			_, err := conn2.ReadFrom(conn1)
			if err != nil {
				break
			}
		}
		wg.Done()
	}(&wg)

	wg.Wait()

	conn1.Close()
	conn2.Close()
}
