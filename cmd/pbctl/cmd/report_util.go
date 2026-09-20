// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package cmd

import reportingpb "github.com/packetbroker/go-api/reporting"

type byToForwarderHomeNetwork []*reportingpb.RoutedMessagesRecord

func (s byToForwarderHomeNetwork) Len() int {
	return len(s)
}

func (s byToForwarderHomeNetwork) Less(i, j int) bool {
	if toI, toJ := s[i].GetTo().AsTime(), s[j].GetTo().AsTime(); toI.Before(toJ) {
		return true
	} else if toI.After(toJ) {
		return false
	}
	if s[i].GetForwarderNetId() < s[j].GetForwarderNetId() {
		return true
	} else if s[i].GetForwarderNetId() > s[j].GetForwarderNetId() {
		return false
	}
	if s[i].GetForwarderTenantId() < s[j].GetForwarderTenantId() {
		return true
	} else if s[i].GetForwarderTenantId() > s[j].GetForwarderTenantId() {
		return false
	}
	if s[i].GetHomeNetworkNetId() < s[j].GetHomeNetworkNetId() {
		return true
	} else if s[i].GetHomeNetworkNetId() > s[j].GetHomeNetworkNetId() {
		return false
	}
	if s[i].GetHomeNetworkTenantId() < s[j].GetHomeNetworkTenantId() {
		return true
	} else if s[i].GetHomeNetworkTenantId() > s[j].GetHomeNetworkTenantId() {
		return false
	}
	return false
}

func (s byToForwarderHomeNetwork) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
