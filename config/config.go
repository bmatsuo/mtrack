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
	"time"

	"github.com/bmatsuo/mtrack/http"
	"github.com/bmatsuo/mtrack/model"
	"github.com/bmatsuo/mtrack/scan"
)

func Configure() error {
	httpaddr := flag.String("http", ":7890", "http server bind address")
	httpstatic := flag.String("http.static", "http/static", "path to mtrack static files")
	dbpath := flag.String("db", "./data/mtrack.sqlite", "sqlite3 database path")
	media := flag.String("media", "", "media directories separated by ':'")
	scandelay := flag.Uint("scan.delay", 5, "minutes between filesystem scans")
	flag.Parse()

	// setup global config
	model.DBPath = *dbpath
	http.HTTPConfig.Addr = *httpaddr
	http.HTTPConfig.StaticPath = *httpstatic
	scandelaydur := time.Duration(*scandelay) * time.Minute
	scanroots := mediaroots(*media)
	scan.Init(scandelaydur, scanroots)
	return nil
}

func mediaroots(env string) []*scan.Root {
	mediapaths := strings.Split(env, ":")
	var roots []*scan.Root
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

		roots = append(roots, &scan.Root{
			Name: name,
			Path: path,
			Exts: []string{".go", ".mp4", ".m4v", ".avi", ".mkv"},
		})
	}
	return roots
}
