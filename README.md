## Fenster
Fenster is a fronted for RDF quad-stores.

It is similar to [Pubby](http://wifo5-03.informatik.uni-mannheim.de/pubby/), but differns in that it shows triples from all public graphs, not just the default graph.

Example of how an resolvable URI is presented in Fenster: [http://data.deichman.no/resource/tnr_1140686](http://data.deichman.no/resource/tnr_1140686)
[screenshot]: https://dl.dropboxusercontent.com/u/27551242/azur.png

### Status
It's an early beta but stable and usefull. Fenster is allready used in production as a frontend for the [RDF-catalogue of Oslo public library](http://data.deichman.no).

### Deployment
Fenster compiles to a statically linked binary, so deployment couldn't be simpler. It's written in Go, so you'll need the [Go toolchain](http://golang.org) to make it compile:
1. `git clone https://github.com/knakk/fenster`
2. cd fenster
3. `make package`
4. cp `fenster.tar.gz` to your server
5. unpack, adjust settings in `config.ini` and run `fenster`

You have to do the build step (3) manually if your target platform is of a different architecture than your compilation platform.
[See this guide](http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go) if you don't know how to do that.

If you're on Ubuntu, you might want to deploy Fenster as an Upstart service. Example config:
```upstart
description "Fenster"

start on filesystem or runlevel [2345]
stop on runlevel [!2345]

respawn

chdir /path/to/fenster
exec ./fenster
```

### Todo
* Logging
* Test different SPARQL endpoints. Currently only tested against Virtuoso.