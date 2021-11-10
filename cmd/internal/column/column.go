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
	sep                = ", "
	maxDevAddrBlocks   = 3
	maxJoinEUIPrefixes = 2
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
	switch t.DefaultAuthentication.(type) {
	case *packetbroker.Target_PbTokenAuth:
		s += " (with PB token auth)"
	case *packetbroker.Target_BasicAuth_:
		s += " (with HTTP basic auth)"
	case *packetbroker.Target_CustomAuth_:
		s += " (with HTTP custom auth)"
	case *packetbroker.Target_TlsClientAuth:
		s += " (with TLS client auth)"
	}
	if l := len(t.OriginNetIdAuthentication); l > 0 {
		s += fmt.Sprintf(" (+%d with custom origin)", l)
	}
	return s
}

// JoinServerFixedEndpoint prints the target as column field.
type JoinServerFixedEndpoint packetbroker.JoinServerFixedEndpoint

func (t *JoinServerFixedEndpoint) String() string {
	if t == nil {
		return ""
	}
	return (packetbroker.Endpoint{
		TenantID: packetbroker.TenantID{
			NetID: packetbroker.NetID(t.NetId),
			ID:    t.TenantId,
		},
		ClusterID: t.ClusterId,
	}).String()
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

// JoinEUIPrefixes prints JoinEUI prefixes as column field.
type JoinEUIPrefixes []*packetbroker.JoinEUIPrefix

func (bs JoinEUIPrefixes) String() string {
	res := make([]string, 0, maxJoinEUIPrefixes+1)
	for i := 0; i < len(bs) && i < maxJoinEUIPrefixes; i++ {
		var (
			b = bs[i]
			s string
		)
		s = fmt.Sprintf("%016X/%d", b.GetValue(), b.GetLength())
		res = append(res, s)
	}
	if more := len(bs) - maxJoinEUIPrefixes; more > 0 {
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

type sortJoinEUIsByPrefix []*packetbroker.JoinEUIPrefix

func (r sortJoinEUIsByPrefix) Len() int {
	return len(r)
}

func (r sortJoinEUIsByPrefix) Less(i, j int) bool {
	if r[i].GetValue() < r[j].GetValue() {
		return true
	} else if r[i].GetValue() == r[j].GetValue() {
		return r[i].GetLength() < r[j].GetLength()
	}
	return false
}

func (r sortJoinEUIsByPrefix) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// WriteJoinEUIPrefixes writes the JoinEUI prefixes as a table.
func WriteJoinEUIPrefixes(w io.Writer, prefixes []*packetbroker.JoinEUIPrefix) error {
	sort.Sort(sortJoinEUIsByPrefix(prefixes))
	for _, b := range prefixes {
		if _, err := fmt.Fprintf(w, "%016X/%d\t\n",
			b.GetValue(),
			b.GetLength(),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeContactInfo(w io.Writer, name string, contactInfo *packetbroker.ContactInfo) error {
	if contactInfo == nil {
		return nil
	}
	return WriteKV(w,
		fmt.Sprintf("%s Name", name), contactInfo.Name,
		fmt.Sprintf("%s Email", name), contactInfo.Email,
		fmt.Sprintf("%s URL", name), contactInfo.Url,
	)
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

	targetAuth := func(auth *packetbroker.Target_Authentication) string {
		switch a := auth.GetValue().(type) {
		case *packetbroker.Target_Authentication_PbTokenAuth:
			return "Packet Broker token"
		case *packetbroker.Target_Authentication_BasicAuth:
			return fmt.Sprintf("Basic username %q, password %q", a.BasicAuth.Username, a.BasicAuth.Password)
		case *packetbroker.Target_Authentication_CustomAuth:
			return a.CustomAuth.Value
		case *packetbroker.Target_Authentication_TlsClientAuth:
			sub, err := x509SubjectFromPair(a.TlsClientAuth.Cert, a.TlsClientAuth.Key)
			if err != nil {
				return err.Error()
			}
			return sub
		default:
			return ""
		}
	}

	var auth string
	switch a := target.DefaultAuthentication.(type) {
	case *packetbroker.Target_PbTokenAuth:
		auth = targetAuth(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_PbTokenAuth{},
		})
	case *packetbroker.Target_BasicAuth_:
		auth = targetAuth(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_BasicAuth{
				BasicAuth: a.BasicAuth,
			},
		})
	case *packetbroker.Target_CustomAuth_:
		auth = targetAuth(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_CustomAuth{
				CustomAuth: a.CustomAuth,
			},
		})
	case *packetbroker.Target_TlsClientAuth:
		auth = targetAuth(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_TlsClientAuth{
				TlsClientAuth: a.TlsClientAuth,
			},
		})
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

	WriteKV(w,
		"Target Protocol", target.Protocol.String(),
		"Target Address", target.Address,
		"Target fNS Path", target.FNsPath,
		"Target sNS Path", target.SNsPath,
		"Target hNS Path", target.HNsPath,
	)
	for i, r := range rootCAs {
		WriteKV(w,
			fmt.Sprintf("Target Root CA #%d", i+1), r,
		)
	}
	WriteKV(w,
		"Target Authorization", auth,
	)
	// TODO: Sort by NetID.
	for netID, auth := range target.OriginNetIdAuthentication {
		WriteKV(w,
			fmt.Sprintf("Target Authorization %s", packetbroker.NetID(netID)), targetAuth(auth),
		)
	}
	return nil
}

// WriteJoinServer writes the Join Server.
func WriteJoinServer(w io.Writer, js *packetbroker.JoinServer) error {
	WriteKV(w,
		"ID", fmt.Sprintf("%d", js.Id),
		"Name", js.GetName(),
	)
	if err := writeContactInfo(w, "Administrator", js.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(w, "Technical", js.GetTechnicalContact()); err != nil {
		return err
	}
	switch resolver := js.Resolver.(type) {
	case *packetbroker.JoinServer_Fixed:
		WriteKV(w,
			"Fixed NetID", packetbroker.NetID(resolver.Fixed.NetId),
			"Fixed Tenant ID", resolver.Fixed.TenantId,
			"Fixed Cluster ID", resolver.Fixed.ClusterId,
		)
	case *packetbroker.JoinServer_Lookup:
		if err := writeTarget(w, resolver.Lookup); err != nil {
			return err
		}
	}
	fmt.Fprintln(w, "\nJoinEUI Prefixes:")
	return WriteJoinEUIPrefixes(w, js.GetJoinEuiPrefixes())
}

// WriteNetwork writes the Network.
func WriteNetwork(w io.Writer, network *packetbroker.Network) error {
	if err := WriteKV(w,
		"NetID", packetbroker.NetID(network.GetNetId()),
		"Name", network.GetName(),
	); err != nil {
		return err
	}
	if err := writeContactInfo(w, "Administrator", network.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(w, "Technical", network.GetTechnicalContact()); err != nil {
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
	if err := writeContactInfo(w, "Administrator", tenant.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(w, "Technical", tenant.GetAdministrativeContact()); err != nil {
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
			netID, tenantID := p.GetHomeNetworkNetId(), p.GetHomeNetworkTenantId()
			if netID == 0 && tenantID == "" {
				fmt.Fprint(w, "\t\t")
			} else {
				fmt.Fprintf(w, "%s\t%s\t",
					packetbroker.NetID(p.GetHomeNetworkNetId()),
					p.GetHomeNetworkTenantId(),
				)
			}
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

// WriteVisibilities writes the gateway visibilities as a table.
func WriteVisibilities(w io.Writer, defaults bool, visibilities ...*packetbroker.GatewayVisibility) error {
	fmt.Fprint(w, "Forwarder\t\t")
	if !defaults {
		fmt.Fprint(w, "Home Network\t\t")
	}
	fmt.Fprintln(w, "Lo\tAp\tAc\tFt\tCi\tSt\tFp\tPr\t")

	for _, v := range visibilities {
		fmt.Fprintf(w, "%s\t%s\t",
			packetbroker.NetID(v.GetForwarderNetId()),
			v.GetForwarderTenantId(),
		)
		if !defaults {
			fmt.Fprintf(w, "%s\t%s\t",
				packetbroker.NetID(v.GetHomeNetworkNetId()),
				v.GetHomeNetworkTenantId(),
			)
		}
		for _, b := range []bool{
			v.GetLocation(),
			v.GetAntennaPlacement(),
			v.GetAntennaCount(),
			v.GetFineTimestamps(),
			v.GetContactInfo(),
			v.GetStatus(),
			v.GetFrequencyPlan(),
			v.GetPacketRates(),
		} {
			if b {
				fmt.Fprint(w, "x")
			}
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintln(w)
	}
	return nil
}
