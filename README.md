## Fenster
Fenster is a fronted for RDF quad-stores.

Example of how a resolvable URI, [http://data.deichman.no/resource/tnr_1140686](http://data.deichman.no/resource/tnr_1140686), is presented in Fenster:
![screenshot](https://dl.dropboxusercontent.com/u/27551242/azur.png)

It is inspired by, and similar to [Pubby](http://wifo5-03.informatik.uni-mannheim.de/pubby/), but differs in that it shows triples from all public graphs, not just the default graph.

### Status
It's still a very young project, but considered stable. Fenster is allready used in production as a frontend for the [RDF-catalogue](http://data.deichman.no) of [Oslo public library](http://www.deichman.no)

### Deployment
Fenster is written in Go, so you'll need the [Go toolchain](http://golang.org/doc/install) in order to build. It compiles to a statically linked binary, so deployment couldn't be simpler:

1. `git clone https://github.com/knakk/fenster`
2. `cd fenster`
3. `make package`*
4. copy `fenster.tar.gz` to your server
5. unpack, adjust settings in `config.ini` and run `fenster`

*You have to do the build step manually if your target platform is of a different architecture than your compilation platform.
[See this guide](http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go) if you don't know where to start.

If you're on Ubuntu, you might want to deploy Fenster as an Upstart service. Example config:
```upstart
description "Fenster"

start on filesystem or runlevel [2345]
stop on runlevel [!2345]

respawn

chdir /path/to/fenster
exec ./fenster
```

### License
GPLv3

### Todo
* Logging
* Test different SPARQL endpoints (currently only tested against Virtuoso)