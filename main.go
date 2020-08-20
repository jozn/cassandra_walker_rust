package main

import (
	"fmt"
	//"./cwalker"
	"github.com/alexflint/go-arg"
)

const (
	DEFAULT_CLUSTER_ADDRESS = "127.0.0.1"
	DEFAULT_PORT            = 9042
	DEFAULT_GO_PACKAGE_NAEM = "xc"
	DEFAULT_OUTPUT          = "./out"
)

func main() {
	args := &ConfigArgs{}
	arg.MustParse(args)
	if args.Host == "" {
		args.Host = DEFAULT_CLUSTER_ADDRESS
	}

	if args.Port == 0 {
		args.Port = DEFAULT_PORT
	}

	if args.Package == "" {
		args.Package = DEFAULT_GO_PACKAGE_NAEM
	}

	if args.Dir == "" {
		args.Dir = DEFAULT_OUTPUT
	}

	Runner(args)

	fmt.Println(args)
}
