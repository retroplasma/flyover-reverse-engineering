#!/bin/bash
protoc --go_out=../pkg/mps/. *.proto
