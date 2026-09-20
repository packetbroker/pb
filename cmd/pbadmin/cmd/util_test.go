// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strconv"
	"testing"

	packetbroker "github.com/packetbroker/go-api/v3"
)

func TestMergeDevAddrBlocks(t *testing.T) {
	t.Parallel()

	equals := func(x, y *packetbroker.DevAddrBlock) bool {
		return x.GetPrefix().GetValue() == y.GetPrefix().GetValue() &&
			x.GetPrefix().GetLength() == y.GetPrefix().GetLength() &&
			x.GetHomeNetworkClusterId() == y.GetHomeNetworkClusterId()
	}

	for i, set := range []struct {
		current,
		add,
		remove []*packetbroker.DevAddrBlock
	}{
		{
			current: []*packetbroker.DevAddrBlock{
				{
					Prefix: &packetbroker.DevAddrPrefix{
						Value:  0x26000000,
						Length: 24,
					},
					HomeNetworkClusterId: "test-1",
				},
				{
					Prefix: &packetbroker.DevAddrPrefix{
						Value:  0x27123400,
						Length: 24,
					},
					HomeNetworkClusterId: "test-2",
				},
			},
			add: []*packetbroker.DevAddrBlock{
				{
					Prefix: &packetbroker.DevAddrPrefix{
						Value:  0x27123400,
						Length: 24,
					},
					HomeNetworkClusterId: "test-2-updated",
				},
				{
					Prefix: &packetbroker.DevAddrPrefix{
						Value:  0x27432100,
						Length: 24,
					},
					HomeNetworkClusterId: "test-3",
				},
			},
			remove: []*packetbroker.DevAddrBlock{
				{
					Prefix: &packetbroker.DevAddrPrefix{
						Value:  0x26000000,
						Length: 24,
					},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			res := mergeDevAddrBlocks(set.current, set.add, set.remove)
			if len(res) != 2 {
				t.Fatalf("unexpected length %d (expected %d)", len(res), 2)
			}
			if !equals(res[0], &packetbroker.DevAddrBlock{
				Prefix: &packetbroker.DevAddrPrefix{
					Value:  0x27123400,
					Length: 24,
				},
				HomeNetworkClusterId: "test-2-updated",
			}) {
				t.Fatalf("unexpected block at position 0: %s at %s", res[0].GetPrefix(), res[0].GetHomeNetworkClusterId())
			}
			if !equals(res[1], &packetbroker.DevAddrBlock{
				Prefix: &packetbroker.DevAddrPrefix{
					Value:  0x27432100,
					Length: 24,
				},
				HomeNetworkClusterId: "test-3",
			}) {
				t.Fatalf("unexpected block at position 1: %s at %s", res[1].GetPrefix(), res[1].GetHomeNetworkClusterId())
			}
		})
	}
}
