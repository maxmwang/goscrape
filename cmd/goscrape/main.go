package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/maxmwang/goscrape/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	company := flag.String("c", "", "company name")
	add := flag.Bool("a", false, "add if not exists")
	flag.Parse()
	if *company == "" {
		panic("please provide a company name with the -c flag")
	}

	conn, err := grpc.Dial("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := proto.NewScraperClient(conn)

	if *add {
		res, err := client.TryThenAdd(context.Background(), &proto.TryRequest{Company: *company})
		if err != nil {
			panic(err)
		}
		fmt.Printf("[company=%s]: site=%s, count=%d, target=%d, exists=%t, added=%t\n", *company, res.Site, res.Count, res.Target, res.Exists, res.Added)
	} else {
		res, err := client.Try(context.Background(), &proto.TryRequest{Company: *company})
		if err != nil {
			panic(err)
		}
		fmt.Printf("[company=%s]: site=%s, count=%d, target=%d, exists=%t\n", *company, res.Site, res.Count, res.Target, res.Exists)
	}
}
