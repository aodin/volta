package main

import (
	"flag"
	"fmt"
	"github.com/aodin/aspect"
	"github.com/aodin/volta/auth/authdb"
)

// An example init script that will print the SQL for applications.
// Add your own applications and functionality!

var applications = map[string][]*aspect.TableElem{
	"auth": {
		authdb.Users,
		authdb.Sessions,
	},
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Available applications:")
		for key := range applications {
			fmt.Printf(" * %s\n", key)
		}
		return
	}

	var invalid bool
	for _, app := range flag.Args() {
		_, ok := applications[app]
		if !ok {
			fmt.Printf(" * '%s' is not a valid application name\n", app)
			invalid = true
		}
	}

	if invalid {
		return
	}

	for _, app := range flag.Args() {
		tables := applications[app]
		fmt.Printf("-- %s\n", app)
		for _, table := range tables {
			fmt.Println(table.Create())
		}
	}
}
