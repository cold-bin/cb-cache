package main

import (
	"context"
	"fmt"
	"github.com/cold-bin/cb-cache/registry"
	"log"
)

func main() {
	n, err := registry.New(context.Background(), "_cb-cache/", []string{"localhost:2379"})
	if err != nil {
		log.Fatalln(err)
	}
	watch := n.Watch(context.Background())
	fmt.Println(n.GetAddress(context.Background()))
	if err := n.Register(context.Background(), "localhost:8081"); err != nil {
		log.Fatalln(err)
	}
	for event := range watch {
		fmt.Println(event)
	}
}
