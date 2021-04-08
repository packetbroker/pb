// Copyright © 2020 The Things Industries B.V.

package column

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	packetbroker "go.packetbroker.org/api/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sep              = ", "
	maxDevAddrBlocks = 3
)

// YesNo prints the boolean as Yes or No.
type YesNo bool

func (yn YesNo) String() string {
	if yn {
		return "Yes"
	}
	return "No"
}

// Target prints the target as column field.
type Target packetbroker.Target

func (t *Target) String() string {
	if t == nil {
		return ""
	}
	s := t.Protocol.String()
	if t.Address != "" {
		s += fmt.Sprintf(": %s", t.Address)
	}
	switch t.Authorization.(type) {
	case *packetbroker.Target_BasicAuth_:
		s += " (with HTTP basic auth)"
	case *packetbroker.Target_CustomAuth_:
		s += " (with HTTP custom auth)"
	case *packetbroker.Target_TlsClientAuth:
		s += " (with TLS client auth)"
	}
	return s
}

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

type sortBlocksByPrefix []*packetbroker.DevAddrBlock

func (r sortBlocksByPrefix) Len() int {
	return len(r)
}

func (r sortBlocksByPrefix) Less(i, j int) bool {
	if r[i].GetPrefix().GetValue() < r[j].GetPrefix().GetValue() {
		return true
	} else if r[i].GetPrefix().GetValue() == r[j].GetPrefix().GetValue() {
		return r[i].GetPrefix().GetLength() < r[j].GetPrefix().GetLength()
	}
	return false
}

func (r sortBlocksByPrefix) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// WriteDevAddrBlocks writes the DevAddr blocks as a table.
func WriteDevAddrBlocks(w io.Writer, blocks []*packetbroker.DevAddrBlock) error {
	sort.Sort(sortBlocksByPrefix(blocks))
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

func x509SubjectFromPair(certPEMBlock, keyPEMBlock []byte) (string, error) {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return "", err
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return "", err
	}
	return x509Cert.Subject.String(), nil
}

func x509Subject(raw []byte) (string, error) {
	var subject pkix.RDNSequence
	if rest, err := asn1.Unmarshal(raw, &subject); err != nil {
		return "", err
	} else if len(rest) != 0 {
		return "", errors.New("trailing data after X.509 subject")
	}
	var res pkix.Name
	res.FillFromRDNSequence(&subject)
	return res.String(), nil
}

func writeTarget(w io.Writer, target *packetbroker.Target) error {
	if target == nil {
		return nil
	}

	var rootCAs []string
	if cas := target.RootCas; len(cas) > 0 {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cas)
		rawSubjects := certPool.Subjects()
		rootCAs = make([]string, len(rawSubjects))
		for i, raw := range rawSubjects {
			subject, err := x509Subject(raw)
			if err != nil {
				rootCAs[i] = err.Error()
			} else {
				rootCAs[i] = subject
			}
		}
	}

	var auth string
	switch a := target.Authorization.(type) {
	case *packetbroker.Target_BasicAuth_:
		auth = fmt.Sprintf("Basic %s:%s", a.BasicAuth.Username, a.BasicAuth.Password)
	case *packetbroker.Target_CustomAuth_:
		auth = a.CustomAuth.Value
	case *packetbroker.Target_TlsClientAuth:
		var err error
		auth, err = x509SubjectFromPair(a.TlsClientAuth.Cert, a.TlsClientAuth.Key)
		if err != nil {
			auth = err.Error()
		}
	}

	if err := WriteKV(w,
		"Target Protocol", target.Protocol.String(),
		"Target Address", target.Address,
		"Target Authorization", auth,
	); err != nil {
		return err
	}
	for i, r := range rootCAs {
		if err := WriteKV(w,
			fmt.Sprintf("Target Root CA #%d", i+1), r,
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
	if err := writeTarget(w, network.GetTarget()); err != nil {
		return err
	}
	fmt.Fprintln(w, "\nDevAddr Blocks:")
	return WriteDevAddrBlocks(w, network.GetDevAddrBlocks())
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
	if err := writeTarget(w, tenant.GetTarget()); err != nil {
		return err
	}
	fmt.Fprintln(w, "\nDevAddr Blocks:")
	return WriteDevAddrBlocks(w, tenant.GetDevAddrBlocks())
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

// Rights formats the API key rights.
type Rights []packetbroker.Right

func (r Rights) String() string {
	rights := make([]string, 0, len(r))
	for _, v := range r {
		rights = append(rights, v.String())
	}
	sort.Strings(rights)
	return strings.Join(rights, ",")
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
