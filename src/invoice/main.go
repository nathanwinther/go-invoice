package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

import (
	"invoice/server"
	"invoice/table"
)

func main() {
	var help bool
	flag.BoolVar(&help, "help", false, "Print help")

	var host string
	flag.StringVar(&host, "host", "", "Host to listen on")

	var port string
	flag.StringVar(&port, "port", "3000", "Port to listen on")

	var root string
	flag.StringVar(&root, "root", ".", "Root directory")

	var start bool
	flag.BoolVar(&start, "start", false, "Start server")

	var create bool
	flag.BoolVar(&create, "create", false, "Create dynamodb tables")

	var desc bool
	flag.BoolVar(&desc, "desc", false, "Description dynamodb tables")

	var del bool
	flag.BoolVar(&del, "del", false, "Delete dynamodb tables")

	flag.Parse()

	err := os.Chdir(root)
	if err != nil {
		panic(err)
	}

	switch {
	case start:
		server.Run(host, port)
		return
	case create:
		table.Create()
		return
	case desc:
		table.Describe()
		return
	case del:
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Are you sure? (Y/n)")
		result, _ := reader.ReadString('\n')
		result = strings.Trim(result, " \t\r\n")
		if result == "Y" {
			table.Delete()
		}
		return
	default:
		flag.PrintDefaults()
		return
	}
}
