package main

import (
	"fmt"
	"os"

	"github.com/zillolo/dht"
)

func main() {

	var host string
	if len(os.Args) > 1 {
		host = "localhost:7000"

		node, err := dht.NewNode(host, dht.LEADER)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		node.Run()
	} else {
		host = "localhost:8000"

		node, err := dht.NewNode(host, dht.FOLLOWER)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		if err := node.Register("localhost:7000"); err != nil {
			fmt.Printf("%v\n", err)
		}
		node.Run()
	}
}
