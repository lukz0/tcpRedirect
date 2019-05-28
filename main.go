package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	selectedSourcePort, selectedTargetPort := getPortsFromUser()
	selectedSourcePortString := strconv.FormatUint(selectedSourcePort, 10)
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	var (
		resolvedSourceAddresses []*net.TCPAddr
		resolvedAddressExists   bool
	)
	for _, a := range addresses {
		a_string := strings.Split(a.String(), "/")[0]
		if strings.ContainsRune(a_string, rune('.')) {
			tcpSourceAddr, err := net.ResolveTCPAddr("tcp", strings.Join([]string{a_string, ":", selectedSourcePortString}, ""))
			if err == nil {
				resolvedAddressExists = true
				resolvedSourceAddresses = append(resolvedSourceAddresses, tcpSourceAddr)
			}
		} else {
			tcpSourceAddr, err := net.ResolveTCPAddr("tcp", strings.Join([]string{"[", a_string, "]:", selectedSourcePortString}, ""))
			if err == nil {
				resolvedAddressExists = true
				resolvedSourceAddresses = append(resolvedSourceAddresses, tcpSourceAddr)
			}
		}
	}
	if !resolvedAddressExists {
		panic("Failed to resolve source address")
	}
	tcpTargetAddr, err := net.ResolveTCPAddr("tcp", strings.Join([]string{"127.0.0.1:", strconv.FormatUint(selectedTargetPort, 10)}, ""))
	if err != nil {
		os.Stdout.Write([]byte("Error resolving target address\n"))
		panic(err)
	}

	var wg sync.WaitGroup
	for _, tcpSourceAddr := range resolvedSourceAddresses {
		tcpListener, err := net.ListenTCP("tcp", tcpSourceAddr)
		if err != nil {
			//os.Stdout.Write([]byte(strings.Join([]string{"Error listening to address: ", tcpSourceAddr.String(), "\n"}, "")))
		} else {
			wg.Add(1)
			os.Stdout.Write([]byte(strings.Join([]string{"Listening to address: ", tcpSourceAddr.String(), "\n"}, "")))
			go func(listener *net.TCPListener, wg_ptr *sync.WaitGroup) {
				for {
					conn, err := listener.Accept()
					if err != nil {
						wg_ptr.Done()
						panic(err)
					}
					go handleConn(conn, tcpTargetAddr)
				}
			}(tcpListener, &wg)
		}
	}
	os.Stdout.Write([]byte(strings.Join([]string{"Redirecting to ", tcpTargetAddr.String(), " Press \"Ctrl + c\" to exit\n"}, "")))
	wg.Wait()
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
