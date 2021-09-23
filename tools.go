// Copyright © 2020 The Things Industries B.V.

//go:build tools

package main

import (
	_ "golang.org/x/lint/golint"
	_ "mvdan.cc/gofumpt/gofumports"
)
