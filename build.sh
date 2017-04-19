#!/bin/sh

if [ "${1}" != "--no-gulp" ]; then
  gulp
fi

export GOPATH=`pwd`

echo Get dependencies...
go get -u github.com/aws/aws-sdk-go

go get -u github.com/nathanwinther/totp
go get -u github.com/nathanwinther/go-uuid4

echo Format...
gofmt -w src/invoice

echo Compile...
go install -v invoice

