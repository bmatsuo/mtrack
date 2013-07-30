// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// config.go [created: Mon, 29 Jul 2013]

// Package config does ....
package config

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmatsuo/mtrack/http"
	"github.com/bmatsuo/mtrack/model"
	"github.com/bmatsuo/mtrack/scan"
)

func Configure() error {
	httpaddr := flag.String("http", ":7890", "http server bind address")
	dbpath := flag.String("db", "./data/mtrack.sqlite", "sqlite3 database path")
	media := flag.String("media", "", "media directories separated by ':'")
	flag.Parse()

	// setup global config
	model.DBPath = *dbpath
	http.HTTPConfig.Addr = *httpaddr
	mediapaths := strings.Split(*media, ":")
	for _, path := range mediapaths {
		var name string

		pieces := strings.SplitN(path, "=", 2)
		if len(pieces) > 1 {
			name, path = pieces[0], pieces[1]
		}

		if path == "" {
			if name != "" {
				log.Print("%s: path missing", name)
			}
			continue
		}

		if name == "" {
			name = filepath.Base(path)
		}

		scan.MediaRoots = append(scan.MediaRoots, scan.FSRoot{name, path})
	}

	return nil
}
