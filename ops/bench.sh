#!/usr/bin/env bash
cd ../tests
go generate
go test -bench=. -benchmem .
