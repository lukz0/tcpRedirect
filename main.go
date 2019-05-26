package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	selectedSourcePort, selectedTargetPort := getPortsFromUser()
	selectedSourceAddr := strings.Join([]string{"127.0.0.1:", strconv.FormatUint(selectedSourcePort, 10)}, "")
	selectedTargetAddr := strings.Join([]string{"127.0.0.1:", strconv.FormatUint(selectedTargetPort, 10)}, "")

	tcpSourceAddr, err := net.ResolveTCPAddr("tcp", selectedSourceAddr)
	if err != nil {
		os.Stdout.Write([]byte("Error resolving source address\n"))
		panic(err)
	}
	tcpTargetAddr, err := net.ResolveTCPAddr("tcp", selectedTargetAddr)
	if err != nil {
		os.Stdout.Write([]byte("Error resolving target address\n"))
		panic(err)
	}

	tcpListener, err := net.ListenTCP("tcp", tcpSourceAddr)
	if err != nil {
		os.Stdout.Write([]byte("Error listening to source address\n"))
		panic(err)
	}

	os.Stdout.Write([]byte("Redirecting... Press \"Ctrl + c\" to exit\n"))
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn, tcpTargetAddr)
	}
}

func handleConn(conn net.Conn, tcpTargetAddr *net.TCPAddr) {
	defer conn.Close()
	targetConn, err := net.DialTCP("tcp", nil, tcpTargetAddr)
	if err != nil {
		panic(err)
	}
	defer targetConn.Close()

	var (
		chanSource chan []byte = chanFromConn(conn)
		chanTarget chan []byte = chanFromConn(targetConn)
	)

	var connectionClosed bool
	for !connectionClosed {
		select {
		case b1 := <-chanSource:
			if b1 == nil {
				connectionClosed = true
			} else {
				targetConn.Write(b1)
			}
		case b2 := <-chanTarget:
			if b2 == nil {
				connectionClosed = true
			} else {
				conn.Write(b2)
			}
		}
	}
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)
		for {
			n, err := conn.Read(b)
			if err != nil {
				c <- nil
				break
			} else if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				c <- res
			}
		}
	}()

	return c
}

func getPortsFromUser() (selectedSourcePort, selectedTargetPort uint64) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	var portSelected bool
	for !portSelected {
		os.Stdout.Write([]byte("Select a port to redirect from:\n"))
		scanner.Scan()
		var (
			input string = strings.TrimSpace(scanner.Text())
			err   error
		)
		if selectedSourcePort, err = strconv.ParseUint(input, 10, 64); err == nil {
			portSelected = true
		} else {
			os.Stdout.Write([]byte("Please enter an valid unsigned integer\n"))
		}
	}

	portSelected = false
	for !portSelected {
		os.Stdout.Write([]byte("Select a port to redirect to:\n"))
		scanner.Scan()
		var (
			input string = strings.TrimSpace(scanner.Text())
			err   error
		)
		if selectedTargetPort, err = strconv.ParseUint(input, 10, 64); err == nil {
			if selectedTargetPort == selectedSourcePort {
				os.Stdout.Write([]byte("Port already selected for input\n"))
			} else {
				portSelected = true
			}
		} else {
			os.Stdout.Write([]byte("Please enter a valid unsigned integer\n"))
		}
	}
	return
}
