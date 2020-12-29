// Copyright © 2020 The Things Industries B.V.

package column

import (
	"fmt"
	"io"
	"strings"
	"time"

	packetbroker "go.packetbroker.org/api/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sep              = ", "
	maxDevAddrBlocks = 3
)

// DevAddrBlocks prints DevAddr blocks as column field.
type DevAddrBlocks []*packetbroker.DevAddrBlock

func (bs DevAddrBlocks) String() string {
	res := make([]string, 0, maxDevAddrBlocks+1)
	for i := 0; i < len(bs) && i < maxDevAddrBlocks; i++ {
		var (
			b = bs[i]
			s string
		)
		if b.GetHomeNetworkClusterId() != "" {
			s = fmt.Sprintf("%08X/%d (%s)", b.GetPrefix().GetValue(), b.GetPrefix().GetLength(), b.HomeNetworkClusterId)
		} else {
			s = fmt.Sprintf("%08X/%d", b.GetPrefix().GetValue(), b.GetPrefix().GetLength())
		}
		res = append(res, s)
	}
	if more := len(bs) - maxDevAddrBlocks; more > 0 {
		res = append(res, fmt.Sprintf("+%d", more))
	}
	return strings.Join(res, sep)
}

// WriteKV writes the key/value pairs.
func WriteKV(w io.Writer, kv ...interface{}) error {
	for i := 0; i < len(kv); i += 2 {
		if _, err := fmt.Fprintf(w, "%v:\t%v\t\n", kv[i], kv[i+1]); err != nil {
			return err
		}
	}
	return nil
}

// WriteDevAddrBlocks writes the DevAddr blocks as a table.
func WriteDevAddrBlocks(w io.Writer, blocks []*packetbroker.DevAddrBlock) error {
	for _, b := range blocks {
		if _, err := fmt.Fprintf(w, "%08X/%d\t%s\t\n",
			b.GetPrefix().GetValue(),
			b.GetPrefix().GetLength(),
			b.GetHomeNetworkClusterId(),
		); err != nil {
			return err
		}
	}
	return nil
}

// WriteNetwork writes the Network.
func WriteNetwork(w io.Writer, network *packetbroker.Network) error {
	if err := WriteKV(w,
		"NetID", packetbroker.NetID(network.GetNetId()),
		"Name", network.GetName(),
	); err != nil {
		return err
	}
	fmt.Fprintln(w, "\nDevAddr Blocks:")
	return WriteDevAddrBlocks(w, network.DevAddrBlocks)
}

// WriteTenant writes the Tenant.
func WriteTenant(w io.Writer, tenant *packetbroker.Tenant) error {
	if err := WriteKV(w,
		"NetID", packetbroker.NetID(tenant.GetNetId()),
		"Tenant ID", tenant.GetTenantId(),
		"Name", tenant.GetName(),
	); err != nil {
		return err
	}
	fmt.Fprintln(w, "\nDevAddr Blocks:")
	return WriteDevAddrBlocks(w, tenant.DevAddrBlocks)
}

// TimeSince formats the timestamp as duration since then, in seconds.
type TimeSince timestamppb.Timestamp

func (t *TimeSince) String() string {
	tmst := (*timestamppb.Timestamp)(t)
	if !tmst.IsValid() {
		return "never"
	}
	d := time.Since(tmst.AsTime())
	if d < 0 {
		d = 0
	}
	d -= d % time.Second
	return d.String()
}

// WritePolicies writes the policies as a table.
func WritePolicies(w io.Writer, defaults bool, policies ...*packetbroker.RoutingPolicy) error {
	fmt.Fprint(w, "Forwarder\t\t")
	if !defaults {
		fmt.Fprint(w, "Home Network\t\t")
	}
	fmt.Fprintln(w, "J\tM\tA\tS\tL\tJ\tM\tA\t")

	for _, p := range policies {
		fmt.Fprintf(w, "%s\t%s\t",
			packetbroker.NetID(p.GetForwarderNetId()),
			p.GetForwarderTenantId(),
		)
		if !defaults {
			fmt.Fprintf(w, "%s\t%s\t",
				packetbroker.NetID(p.GetHomeNetworkNetId()),
				p.GetHomeNetworkTenantId(),
			)
		}
		for _, b := range []bool{
			p.GetUplink().GetJoinRequest(),
			p.GetUplink().GetMacData(),
			p.GetUplink().GetApplicationData(),
			p.GetUplink().GetLocalization(),
			p.GetUplink().GetSignalQuality(),
		} {
			if b {
				fmt.Fprint(w, "▲")
			}
			fmt.Fprint(w, "\t")
		}
		for _, b := range []bool{
			p.GetDownlink().GetJoinAccept(),
			p.GetDownlink().GetMacData(),
			p.GetDownlink().GetApplicationData(),
		} {
			if b {
				fmt.Fprint(w, "▼")
			}
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintln(w)
	}
	return nil
}
