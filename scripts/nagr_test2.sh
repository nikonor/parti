#!/bin/bash 

wrk -c 200 -d 10s -t 10 -s ./docs/wrk.lua http://localhost:8087/foo
