[godoc.org]: http://godoc.org/github.com/bmatsuo/mtrack/ "godoc.org"

This program is still under development and may change radically without notice.

Install
=======

- Download or build a distribution archive (building is the only option).

- Extract the distribution archive in an appropriate location.
```
    $ sudo mkdir -p /usr/local/share/mtrack
    $ sudo tar -C /usr/local/share/mtrack xvzf ./mtrack-123abcd-myos-myarch.tar.gz
```
- Create a shell script to run the server
```
    $ sudo echo '#!/bin/bash' > /usr/local/bin/mtrack
    $ sudo echo /usr/local/share/mtrack/mtrack-123abc-myos-myarch/mtrack -config /etc/mtrack.toml >> /usr/local/bin/mtrack
    $ sudo chmod +x /usr/local/bin/mtrack
```
- Create a configuration file
```
    $ sudo cat > /etc/mtrack.toml
    [HTTP]
    Bind = ":7890"
    StaticRoot = "/usr/local/share/mtrack/mtrack-123abc-myos-myarch/static"

    [DB]
    Path = "/usr/local/share/mtrack/mtrack.sqlite"

    # ... root definitions (see example.mtrack.toml)
    ^D
```

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
