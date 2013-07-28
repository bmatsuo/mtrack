// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// mtrack.go [created: Sat, 27 Jul 2013]

package main

import (
	"log"
	"flag"
)

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	httpaddr := flag.String("http", ":7890", "http server bind address")
	dbpath := flag.String("db", "./mtrack.sqlite", "sqlite3 database path")
	flag.Parse()

	DBPath = *dbpath
	Check(DBInit())

	HTTPConfig.Addr = *httpaddr
	Check(HTTPStart())
}
