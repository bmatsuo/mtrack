// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// static.go [created: Sat,  3 Aug 2013]

package http

import (
	"net/http"
)

func FileServer() http.Handler {
	return http.FileServer(http.Dir(HTTPConfig.StaticPath))
}
