package main

import (
	"fmt"
	"os"

	"github.com/zillolo/dht"
)

func main() {

	if len(os.Args) > 4 {
		return
	} else if len(os.Args) == 4 {
		if os.Args[3] == "--leader" {
			node, err := dht.NewNode(os.Args[2], dht.LEADER)
			if err != nil {
				fmt.Printf("Error during node creation.\n")
				return
			}
			go node.Run()

			InitAPI(os.Args[1], node)
		}
	} else if len(os.Args) == 3 {
		node, err := dht.NewNode(os.Args[2], dht.FOLLOWER)
		if err != nil {
			fmt.Printf("Error during node creation.\n")
			return
		}
		node.Register(os.Args[1])
		node.Run()
	}
}
