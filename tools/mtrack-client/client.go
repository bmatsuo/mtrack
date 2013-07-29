// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// client.go [created: Mon, 29 Jul 2013]

/*
a command line tool for accessing the mtrack API.

	mtrack-client [FLAGS] ENDPOINT [JSON]
	mtrack-client [FLAGS] -write
	mtrack-client -h

If an access token is provided, it is included in the request's Authorization
header. An authorization header is necessary for any POST request.

If a JSON argument is given, the request is made as a POST request. Otherwise
the request method will be GET.

If the -write flag is given, a configuration file is written to PATH
(or ~/.config/mtrack-client.json) with values specified in flags.

Use the -h flag to see a list of available flags.
*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatsuo/jsonutil"
)

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var (
	defaultHttpAddr   = "localhost:7890"
	defaultConfigPath = filepath.Join(
		os.Getenv("HOME"),
		"/.config/mtrack-client.json")
)

var Config struct {
	HTTPAddr    string `json:"http,omitempty"`
	AccessToken string `json:"access,omitempty"`
}

func main() {
	Check(Configure())
	Check(Request(flag.Args()))
}

func Request(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing command argument")
	}

	command, args := args[0], args[1:]
	if command[0] != '/' {
		command = "/" + command
	}

	uri := "http://" + Config.HTTPAddr + command
	method := "GET"
	var body string
	if len(args) > 0 {
		method = "POST"
		body, args = args[0], args[1:]
	}

	req, err := http.NewRequest(method, uri, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if Config.AccessToken != "" {
		req.Header.Set("Authorization", "token "+Config.AccessToken)
	}

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	contentType = strings.SplitN(contentType, ";", 2)[0]
	contentType = strings.Trim(contentType, " ")
	if contentType == "application/json" {
		buf := new(bytes.Buffer)
		err = json.Indent(buf, p, "", "\t")
		fmt.Println(buf.String())
	} else {
		fmt.Println(string(p))
	}

	return nil
}

func Configure() error {
	httpaddr := flag.String("http", "", "http server address")
	access := flag.String("access", "", "access token for authorization")
	noaccess := flag.Bool("noaccess", false, "ignore token in config file")
	configPath := flag.String("config", defaultConfigPath, "config file path")
	writeConfig := flag.Bool("write", false, "write config file from flags")
	flag.Parse()

	_, err := os.Stat(*configPath)
	switch {
	case os.IsNotExist(err):
		break
	case err == nil:
		err := jsonutil.UnmarshalFile(*configPath, &Config)
		if err != nil {
			return err
		}
	default:
	}

	if *noaccess {
		Config.AccessToken = *access
	} else if *access != "" {
		Config.AccessToken = *access
	}

	if *httpaddr != "" {
		Config.HTTPAddr = *httpaddr
	} else if Config.HTTPAddr == "" {
		Config.HTTPAddr = defaultHttpAddr
	}

	if *writeConfig {
		dir := filepath.Dir(*configPath)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		err := jsonutil.MarshalIndentFile(*configPath, 0600, "", "\t", Config)
		if err != nil {
			return err
		}

		os.Exit(0)
	}

	return nil
}
