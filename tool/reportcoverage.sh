#!/bin/bash
scriptdir=`dirname "$(realpath $0)"`
projectdir=`dirname ${scriptdir}`

cd projectdir # working directory
go test -covermode=count -coverprofile=coverage.out
bash <(curl -s https://codecov.io/bash) -t 2add3402-9c35-42d1-b5f7-7e9b791a7750