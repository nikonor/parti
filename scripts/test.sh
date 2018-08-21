#!/bin/bash

dop=""
if [ $# -ne 0 ]; then
    dop="-run "$@
fi 

go test -v *.go $dop
