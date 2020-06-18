package main

import (
	"flag"
	"fmt"
	"os"

	tool "github.com/liuximu/sqlmy/tool"
)

var md = &tool.MysqlDao{
	Output: os.Stderr,
}

var password, user, port, host string

func main() {
	flag.StringVar(&md.DB, "db", "", "db: db name, must set")
	flag.StringVar(&md.Table, "t", "", "table name, if no set, all table in database will be created")
	flag.StringVar(&password, "p", "", "password")
	flag.StringVar(&user, "u", "root", "user")
	flag.StringVar(&port, "P", "3306", "port")
	flag.StringVar(&host, "h", "127.0.0.1", "host name")
	flag.StringVar(&md.PkgName, "pkg", "dao", "package name")
	flag.BoolVar(&md.WithJsonFlag, "json", false, "with json tag")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: dao -db demo_db [-t demo_table] [-u root] [-p password] [-h 127.0.0.1] [-P 3306] [-json false] [-pkg dao]
		
Options:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if md.DB == "" {
		flag.Usage()
		return
	}

	md.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, md.DB)
	md.Create()
}
