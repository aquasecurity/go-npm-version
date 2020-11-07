#!/bin/sh

cd `dirname $0`

node generator/main.js > ../pkg/constraint_testcase.go
