// Copyright © 2021 The Things Industries B.V.

package cmd

import packetbroker "go.packetbroker.org/api/v3"

func mergeDevAddrBlocks(current, add, remove []*packetbroker.DevAddrBlock) []*packetbroker.DevAddrBlock {
	equals := func(x, y *packetbroker.DevAddrBlock) bool {
		return x.Prefix.Value == y.Prefix.Value && x.Prefix.Length == y.Prefix.Length
	}
	for _, a := range add {
		var found bool
		for i, c := range current {
			if equals(a, c) {
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
	for _, c := range current {
		var found bool
		for _, r := range remove {
			if equals(c, r) {
				found = true
				break
			}
		}
		if !found {
			res = append(res, c)
		}
	}
	return res
}
