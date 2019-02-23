#!/bin/bash

rm -Rf ./bin

gox -osarch="linux/amd64" -osarch="linux/386" -osarch="darwin/amd64" -osarch="darwin/386" -osarch="windows/amd64" -osarch="windows/386" -output="./bin/{{.Dir}}_{{.OS}}_{{.Arch}}"

node script.js
rm -Rf /home/akumzy/projects/watcher/bin
cp -r ./bin /home/akumzy/projects/watcher/bin