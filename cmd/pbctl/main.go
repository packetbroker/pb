// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point of the pbctl command-line interface.
package main

import "go.packetbroker.org/pb/cmd/pbctl/cmd"

func main() {
	cmd.Execute()
}
