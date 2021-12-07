// Copyright © 2021 The Things Industries B.V.

package cmd

import reportingpb "go.packetbroker.org/api/reporting"

type byToForwarderHomeNetwork []*reportingpb.RoutedMessagesRecord

func (s byToForwarderHomeNetwork) Len() int {
	return len(s)
}

func (s byToForwarderHomeNetwork) Less(i, j int) bool {
	if toI, toJ := s[i].To.AsTime(), s[j].To.AsTime(); toI.Before(toJ) {
		return true
	} else if toI.After(toJ) {
		return false
	}
	if s[i].ForwarderNetId < s[j].ForwarderNetId {
		return true
	} else if s[i].ForwarderNetId > s[j].ForwarderNetId {
		return false
	}
	if s[i].ForwarderTenantId < s[j].ForwarderTenantId {
		return true
	} else if s[i].ForwarderTenantId > s[j].ForwarderTenantId {
		return false
	}
	if s[i].HomeNetworkNetId < s[j].HomeNetworkNetId {
		return true
	} else if s[i].HomeNetworkNetId > s[j].HomeNetworkNetId {
		return false
	}
	if s[i].HomeNetworkTenantId < s[j].HomeNetworkTenantId {
		return true
	} else if s[i].HomeNetworkTenantId > s[j].HomeNetworkTenantId {
		return false
	}
	return false
}

func (s byToForwarderHomeNetwork) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
