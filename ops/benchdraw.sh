#!/usr/bin/env bash
go generate ../tests
go test -bench=. -benchmem ../... > benchdraw.out.txt
