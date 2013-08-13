// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// config.go [created: Mon, 29 Jul 2013]

// Package config does ....
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bmatsuo/mtrack/http"
	"github.com/bmatsuo/mtrack/model"
	"github.com/bmatsuo/mtrack/scan"
)

var Config = struct {
	HTTP struct {
		Bind       string
		StaticRoot string
	}
	DB struct {
		Path string
	}
	Root  map[string]*scan.Root
	Roots []*scan.Root `toml:"-" json:"-"`
}{}

func loadConfig(path string) (toml.MetaData, error) {
	var zero toml.MetaData
	var err error
	// locate and open the config file. this is funced up
	var f *os.File
	paths := make([]string, 0, 3)
	if path != "" {
		paths = append(paths, path)
	} else {
		home := os.Getenv("HOME")
		if home != "" {
			paths = append(paths, ".config", "mtrack.toml")
		}
		paths = append(paths, "/etc/mtrack.toml")
	}
	for _, path := range paths {
		f, err = os.Open(path)
		log.Printf("%q; %T %v", path, err, err)
		if os.IsNotExist(err) {
			continue
		}
		break
	}
	if os.IsNotExist(err) {
		if path != "" {
			return zero, fmt.Errorf("path does not exist; %v", path)
		}
		return zero, nil
		// default paths not existing is ok.
	} else if err != nil {
		return zero, err
	}

	defer f.Close()
	return toml.DecodeReader(f, &Config)
}

func Configure() error {
	flag.StringVar(&Config.HTTP.Bind, "http", ":7890", "http server bind address")
	flag.StringVar(&Config.DB.Path, "db", "", "sqlite3 database path")
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	_, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if Config.Root == nil {
		Config.Root = make(map[string]*scan.Root)
	}
	Config.Roots = make([]*scan.Root, 0, len(Config.Root))
	for k, root := range Config.Root {
		root.Name = k
		Config.Roots = append(Config.Roots, root)
	}

	if Config.HTTP.Bind == "" {
		return fmt.Errorf("unknown http server bind address")
	}
	if Config.HTTP.StaticRoot == "" {
		return fmt.Errorf("unknown static root directory")
	}
	if Config.DB.Path == "" {
		return fmt.Errorf("unknown database path")
	}

	// setup global config
	model.DBPath = Config.DB.Path
	http.HTTPConfig.Addr = Config.HTTP.Bind
	http.HTTPConfig.StaticPath = Config.HTTP.StaticRoot

	p, err := json.Marshal(Config.Root)

	log.Print(string(p), err)
	scan.Init(5*time.Minute, Config.Roots) // TODO handle per-root delays

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
