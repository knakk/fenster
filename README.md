## Fenster
Fenster is a fronted for RDF quad-stores.

It is inspired by, and similar to [Pubby](http://wifo5-03.informatik.uni-mannheim.de/pubby/), but differs in that it shows triples from all public graphs, not just the default graph.

Example of how a resolvable URI, [http://data.deichman.no/resource/tnr_1140686](http://data.deichman.no/resource/tnr_1140686), is presented in Fenster:
![screenshot](https://dl.dropboxusercontent.com/u/27551242/azur.png)

### Status
Fenster is stable and has been in production since November 2013 as a frontend for the [RDF-catalogue](http://data.deichman.no) of [Oslo public library](http://www.deichman.no). Currently it's only been tested against Virtuoso, but presumably any compliant SPARQL endpoint should work. Please let us know if you run into any issues.

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

#### Apache routing
If Fenster is running on same server as the RDF-store, you'll have to proxy the requests to the SPARQL endpoint.

Here is an example Apache config, given Fenster running on localhost:8080 and SPARQL endpoint running on localhost:8890/sparql:

```apache
<VirtualHost *:80>

    ServerAdmin serveradmin@example.no
    DocumentRoot /var/www/example.com
    ServerName example.com

    ProxyRequests off
    ProxyPreserveHost on
    ProxyTimeout        300
    # Proxy ACL
    <Proxy *>
        Order allow,deny
        Allow from all
    </Proxy>
    RewriteEngine on
    RewriteRule ^/sparql(.*)$ http://localhost:8890/sparql$1 [P]

    # default proxy if not handled above
    ProxyPass / http://example.com:8080/ timeout=300
    ProxyPassReverse / http://example.com:8080/

</VirtualHost>
```


### License
GPLv3
