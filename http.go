// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http.go [created: Sat, 27 Jul 2013]

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var HTTPConfig struct{
	Addr string
}

func HTTPStart() error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(http.NotFound)
	router.Methods("GET").Path("/").HandlerFunc(Index)

	log.Printf("Serving HTTP on at %v", HTTPConfig.Addr)
	return http.ListenAndServe(HTTPConfig.Addr, router)
}

func Index(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "<h1>boom</h1>")

}
