// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package csv writes Packet Broker reports in CSV format.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	reportingpb "go.packetbroker.org/api/reporting"
	packetbroker "go.packetbroker.org/api/v3"
)

type uplinkMessageProcessingError struct {
	code   packetbroker.UplinkMessageProcessingError
	suffix string
}

type uplinkMessageProcessingErrors []uplinkMessageProcessingError

func (s uplinkMessageProcessingErrors) Len() int {
	return len(s)
}

func (s uplinkMessageProcessingErrors) Less(i, j int) bool {
	return s[i].suffix < s[j].suffix
}

func (s uplinkMessageProcessingErrors) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type downlinkMessageProcessingError struct {
	code   packetbroker.DownlinkMessageProcessingError
	suffix string
}

type downlinkMessageProcessingErrors []downlinkMessageProcessingError

func (s downlinkMessageProcessingErrors) Len() int {
	return len(s)
}

func (s downlinkMessageProcessingErrors) Less(i, j int) bool {
	return s[i].suffix < s[j].suffix
}

func (s downlinkMessageProcessingErrors) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func networkOrTenantName(nwk *packetbroker.NetworkOrTenant) string {
	if nwk == nil {
		return ""
	}
	switch n := nwk.GetValue().(type) {
	case *packetbroker.NetworkOrTenant_Network:
		return n.Network.GetName()
	case *packetbroker.NetworkOrTenant_Tenant:
		return n.Tenant.GetName()
	}
	return ""
}

// WriteRoutedMessages writes the records of routed messages in CSV format.
func WriteRoutedMessages(
	out io.Writer,
	records []*reportingpb.RoutedMessagesRecord,
	networks map[packetbroker.TenantID]*packetbroker.NetworkOrTenant,
) error {
	writer := csv.NewWriter(out)

	var (
		uplinkErrs   = make(uplinkMessageProcessingErrors, 0, len(packetbroker.UplinkMessageProcessingError_name))
		downlinkErrs = make(downlinkMessageProcessingErrors, 0, len(packetbroker.DownlinkMessageProcessingError_name))
	)
	for c, n := range packetbroker.UplinkMessageProcessingError_name {
		uplinkErrs = append(uplinkErrs, uplinkMessageProcessingError{
			packetbroker.UplinkMessageProcessingError(c),
			strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(n), "uplink_"), "_error"),
		})
	}
	for c, n := range packetbroker.DownlinkMessageProcessingError_name {
		downlinkErrs = append(downlinkErrs, downlinkMessageProcessingError{
			packetbroker.DownlinkMessageProcessingError(c),
			strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(n), "downlink_"), "_error"),
		})
	}
	sort.Sort(uplinkErrs)
	sort.Sort(downlinkErrs)

	// The fields are: 7 identifying fields, and for each of the 4 message types: routed, success and errors.
	numFields := 7 + 4*2 + 2*len(uplinkErrs) + 2*len(downlinkErrs)
	header := make([]string, 0, numFields)
	header = append(header,
		"date",
		"forwarder_net_id",
		"forwarder_tenant_id",
		"forwarder_name",
		"home_network_net_id",
		"home_network_tenant_id",
		"home_network_name",
	)
	header = append(header,
		"uplink_join_routed",
		"uplink_join_processed_success",
	)
	for _, e := range uplinkErrs {
		header = append(header, fmt.Sprintf("uplink_join_processed_error_%s", e.suffix))
	}
	header = append(header,
		"uplink_data_routed",
		"uplink_data_processed_success",
	)
	for _, e := range uplinkErrs {
		header = append(header, fmt.Sprintf("uplink_data_processed_error_%s", e.suffix))
	}
	header = append(header,
		"downlink_join_routed",
		"downlink_join_processed_success",
	)
	for _, e := range downlinkErrs {
		header = append(header, fmt.Sprintf("downlink_join_processed_error_%s", e.suffix))
	}
	header = append(header,
		"downlink_data_routed",
		"downlink_data_processed_success",
	)
	for _, e := range downlinkErrs {
		header = append(header, fmt.Sprintf("downlink_data_processed_error_%s", e.suffix))
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("csv: write header: %w", err)
	}

	for _, rec := range records {
		uplinkErrFields := make([]string, len(uplinkErrs))
		downlinkErrFields := make([]string, len(downlinkErrs))

		row := make([]string, 0, numFields)
		row = append(row,
			rec.GetTo().AsTime().Format("2006-01-02"),
			packetbroker.NetID(rec.GetForwarderNetId()).String(),
			rec.GetForwarderTenantId(),
			networkOrTenantName(networks[packetbroker.ForwarderTenantID(rec)]),
			packetbroker.NetID(rec.GetHomeNetworkNetId()).String(),
			rec.GetHomeNetworkTenantId(),
			networkOrTenantName(networks[packetbroker.HomeNetworkTenantID(rec)]),
		)
		row = append(row,
			strconv.FormatUint(rec.GetUplink().GetJoinRequestsRoutedCount(), 10),
			strconv.FormatUint(rec.GetUplink().GetJoinRequestsProcessedSuccessCount(), 10),
		)
		for i, e := range uplinkErrs {
			uplinkErrFields[i] = "0"
			for _, u := range rec.GetUplink().GetJoinRequestsProcessedErrorCount() {
				if u.GetErrorType() == e.code {
					uplinkErrFields[i] = strconv.FormatUint(u.GetCount(), 10)
					break
				}
			}
		}
		row = append(row, uplinkErrFields...)
		row = append(row,
			strconv.FormatUint(rec.GetUplink().GetDataMessagesRoutedCount(), 10),
			strconv.FormatUint(rec.GetUplink().GetDataMessagesProcessedSuccessCount(), 10),
		)
		for i, e := range uplinkErrs {
			uplinkErrFields[i] = "0"
			for _, u := range rec.GetUplink().GetDataMessagesProcessedErrorCount() {
				if u.GetErrorType() == e.code {
					uplinkErrFields[i] = strconv.FormatUint(u.GetCount(), 10)
					break
				}
			}
		}
		row = append(row, uplinkErrFields...)
		row = append(row,
			strconv.FormatUint(rec.GetDownlink().GetJoinAcceptsRoutedCount(), 10),
			strconv.FormatUint(rec.GetDownlink().GetJoinAcceptsProcessedSuccessCount(), 10),
		)
		for i, e := range downlinkErrs {
			downlinkErrFields[i] = "0"
			for _, u := range rec.GetDownlink().GetJoinAcceptsProcessedErrorCount() {
				if u.GetErrorType() == e.code {
					downlinkErrFields[i] = strconv.FormatUint(u.GetCount(), 10)
					break
				}
			}
		}
		row = append(row, downlinkErrFields...)
		row = append(row,
			strconv.FormatUint(rec.GetDownlink().GetDataMessagesRoutedCount(), 10),
			strconv.FormatUint(rec.GetDownlink().GetDataMessagesProcessedSuccessCount(), 10),
		)
		for i, e := range downlinkErrs {
			downlinkErrFields[i] = "0"
			for _, u := range rec.GetDownlink().GetDataMessagesProcessedErrorCount() {
				if u.GetErrorType() == e.code {
					downlinkErrFields[i] = strconv.FormatUint(u.GetCount(), 10)
					break
				}
			}
		}
		row = append(row, downlinkErrFields...)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("csv: write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("csv: write: %w", err)
	}
	return nil
}
