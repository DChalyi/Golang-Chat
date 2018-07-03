package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

const connprot = "tcp"
type ConsoleClearer struct {

	deleteChars int
	mux sync.Mutex
}
var delstr string

func main() {

clearer:=&ConsoleClearer{deleteChars:0}
	connection, err := net.Dial(connprot, os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for {
			response := make([]byte, 1024)
			n, err := connection.Read(response)
			if err == io.EOF {
				fmt.Println("Server stopped !")
				fmt.Println("Exit...")
				os.Exit(1)
			}

			fmt.Print(delstr + string(response[:n]))
		}

	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		clearer.deleteChars = len(text)+1

		for ;clearer.deleteChars>0;clearer.deleteChars--{
			delstr+=fmt.Sprintf("%c", 8)
		}
		del:=bufio.NewWriter(os.Stdout)
		_, err= del.Write([]byte(delstr))


		if err != nil {
			fmt.Println(err)
		}
		connection.Write([]byte(text))
	}
}
