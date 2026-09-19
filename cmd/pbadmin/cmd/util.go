// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import packetbroker "go.packetbroker.org/api/v3"

func mergeDevAddrBlocks(current, add, remove []*packetbroker.DevAddrBlock) []*packetbroker.DevAddrBlock {
	equals := func(x, y *packetbroker.DevAddrBlock) bool {
		return x.GetPrefix().GetValue() == y.GetPrefix().GetValue() &&
			x.GetPrefix().GetLength() == y.GetPrefix().GetLength()
	}
	for _, a := range add {
		var found bool
		for i, block := range current {
			if equals(a, block) {
				found = true
				current[i] = a
				break
			}
		}
		if !found {
			current = append(current, a)
		}
	}
	res := make([]*packetbroker.DevAddrBlock, 0, len(current)+len(add)-len(remove))
	for _, block := range current {
		var found bool
		for _, removed := range remove {
			if equals(block, removed) {
				found = true
				break
			}
		}
		if !found {
			res = append(res, block)
		}
	}
	return res
}
