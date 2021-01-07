package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/gocql/gocql"
)

const (
	DEFAULT_CLUSTER_ADDRESS = "127.0.0.1"
	DEFAULT_PORT            = 9042
	DEFAULT_GO_PACKAGE_NAEM = "xc"
	DEFAULT_OUTPUT          = "./out"
)

type ConfigArgs struct {
	Keyspaces []string `arg:"positional,-k,help:cassandra keyspaces to buildGo "`
	Host      string   `arg:"-c,help:cassandra cluster address (default 127.0.0.1)"`
	Port      int      `arg:"-p,help:cassandra port (default 9042)"`
	Verbose   bool     `arg:"-v,help:verbosity Log"`
	Dir       string   `arg:"-d,help:output of generated codes (default './')"`
	Package   string   `arg:"help:package of go"`
	Minimize  bool     `arg:"-m,help:minimize docs"`
}

var gen = &GenOut{}
var args *ConfigArgs

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

	runner(args)

	fmt.Println(args)
}

func runner(arg *ConfigArgs) {
	args = arg
	gen.Package = args.Package

	for _, db := range arg.Keyspaces {
		// connect to the cluster
		cluster := gocql.NewCluster(arg.Host)
		cluster.Keyspace = db
		cluster.Consistency = gocql.One
		session, err := cluster.CreateSession()
		NoErr(err)
		defer session.Close()

		tables := loadTables(db, cluster)

		loadColumns(tables, cluster)

		for _, t := range tables {
			gen.TablesExtracted = append(gen.TablesExtracted, t)
		}
	}

	setTableParams(gen)

	fmt.Println("==========================")
	PertyPrint(gen)

	//buildGo(gen)
	buildRust(gen)
}
