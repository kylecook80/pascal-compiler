#!/bin/bash

rm *.txt
go build
./compiler test.pas
