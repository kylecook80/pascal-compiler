#!/bin/bash

rm *.txt
go build
./compiler src_with_errors.pas
