#!/bin/bash
protoc --go_out=../pkg/fly/internal/pro/. *.proto
