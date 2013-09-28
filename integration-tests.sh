#!/bin/sh
./fenster & echo $! > fenster.pid
sleep 1 # give the server time to start
casperjs test browser_test.js --no-colors
kill $(cat fenster.pid)
rm fenster.pid