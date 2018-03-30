#!/bin/bash

go get golang.org/x/oauth2/google
go get google.golang.org/api/cloudbilling/v1
go get google.golang.org/api/cloudresourcemanager/v1

gofmt -s -w *.go

go test

go install
