[godoc.org]: http://godoc.org/github.com/bmatsuo/mtrack/ "godoc.org"

Install
=======

- Download or build a distribution archive (building is the only option).
- Extract the distribution archive in an appropriate location.
    $ sudo mkdir -p /usr/local/share/mtrack
    $ sudo tar -C /usr/local/share/mtrack xvzf ./mtrack-123abcd.tar.gz
- Create a shell script to run the server
    $ echo '#!/bin/bash' > /usr/local/bin/mtrack
    $ echo /usr/local/share/mtrack/mtrack-123abc/mtrack -config /etc/mtrack.toml >> /usr/local/bin/mtrack
- Create a configuration file
    $ sudo cat > /etc/.mtrack.toml
    [HTTP]
    Bind = ":7890"
    StaticRoot = "/usr/local/share/mtrack/mtrack-123abc/static"

    [DB]
    Path = "/usr/local/share/mtrack/data/mtrack.sqlite"

    # ... root definitions (see example.mtrack.toml)
    ^D

Run the server
==============

    git clone git@github.com:bmatsuo/mtrack.git $GOPATH/github.com/bmatsuo/mtrack
    cd $GOPATH/github.com/bmatsuo/mtrack
    make start-dist

then open [the web app](http://localhost:7890)

Docs
====

On [godoc.org][]

Author
======

Bryan Matsuo [bryan dot matsuo at gmail dot com]

Copyright & License
===================

Copyright (c) 2013, Bryan Matsuo.
All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
