#!/bin/bash

rm *.txt
go build
./compiler final_src.pas
