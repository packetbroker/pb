// SPDX-FileCopyrightText: Copyright 2020 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package column provides formatting of Packet Broker entities as columns and tables for the CLI.
package column

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	packetbroker "go.packetbroker.org/api/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sep                = ", "
	maxDevAddrBlocks   = 1
	maxJoinEUIPrefixes = 2
)

// Writer writes tabulated command output.
// The first write error is recorded and returned by Flush, so that output can be written without checking each write.
type Writer struct {
	tab *tabwriter.Writer
	err error
}

// NewWriter returns a new Writer that writes tabulated output to out.
func NewWriter(out io.Writer) *Writer {
	return &Writer{
		tab: tabwriter.NewWriter(out, 0, 0, 3, ' ', 0),
	}
}

// Write implements io.Writer. After a write error, subsequent writes are discarded and return the recorded error.
func (w *Writer) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.tab.Write(p)
	if err != nil {
		w.err = fmt.Errorf("write output: %w", err)
		return n, w.err
	}
	return n, nil
}

// Println writes the line, followed by a newline. Write errors are returned by Flush.
func (w *Writer) Println(line string) {
	// The error is recorded by Write.
	_, _ = fmt.Fprintln(w, line)
}

// Printf writes the formatted output. Write errors are returned by Flush.
func (w *Writer) Printf(format string, args ...any) {
	// The error is recorded by Write.
	_, _ = fmt.Fprintf(w, format, args...)
}

// Flush flushes the buffered output and returns the first write error, if any.
func (w *Writer) Flush() error {
	if w.err != nil {
		return w.err
	}
	if err := w.tab.Flush(); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

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
	str := t.Protocol.String()
	if t.Address != "" {
		str += fmt.Sprintf(": %s", t.Address)
	}
	switch t.DefaultAuthentication.(type) {
	case *packetbroker.Target_PbTokenAuth:
		str += " (with PB token auth)"
	case *packetbroker.Target_BasicAuth_:
		str += " (with HTTP basic auth)"
	case *packetbroker.Target_CustomAuth_:
		str += " (with HTTP custom auth)"
	case *packetbroker.Target_TlsClientAuth:
		str += " (with TLS client auth)"
	}
	if l := len(t.OriginNetIdAuthentication); l > 0 {
		str += fmt.Sprintf(" (+%d with custom origin)", l)
	}
	return str
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
		b := bs[i]
		if b.GetHomeNetworkClusterId() != "" {
			res = append(res, fmt.Sprintf("%08X/%d (%s)",
				b.GetPrefix().GetValue(), b.GetPrefix().GetLength(), b.GetHomeNetworkClusterId(),
			))
		} else {
			res = append(res, fmt.Sprintf("%08X/%d", b.GetPrefix().GetValue(), b.GetPrefix().GetLength()))
		}
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
		b := bs[i]
		res = append(res, fmt.Sprintf("%016X/%d", b.GetValue(), b.GetLength()))
	}
	if more := len(bs) - maxJoinEUIPrefixes; more > 0 {
		res = append(res, fmt.Sprintf("+%d", more))
	}
	return strings.Join(res, sep)
}

// WriteKV writes the key/value pairs.
func WriteKV(out io.Writer, kv ...string) error {
	for i := 0; i+1 < len(kv); i += 2 {
		if _, err := fmt.Fprintf(out, "%v:\t%v\t\n", kv[i], kv[i+1]); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

// writeLine writes the line to out, followed by a newline.
func writeLine(out io.Writer, line string) error {
	if _, err := fmt.Fprintln(out, line); err != nil {
		return fmt.Errorf("write output: %w", err)
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
func WriteDevAddrBlocks(out io.Writer, blocks []*packetbroker.DevAddrBlock) error {
	sort.Sort(sortBlocksByPrefix(blocks))
	for _, b := range blocks {
		if _, err := fmt.Fprintf(out, "%08X/%d\t%s\t\n",
			b.GetPrefix().GetValue(),
			b.GetPrefix().GetLength(),
			b.GetHomeNetworkClusterId(),
		); err != nil {
			return fmt.Errorf("write output: %w", err)
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
func WriteJoinEUIPrefixes(out io.Writer, prefixes []*packetbroker.JoinEUIPrefix) error {
	sort.Sort(sortJoinEUIsByPrefix(prefixes))
	for _, b := range prefixes {
		if _, err := fmt.Fprintf(out, "%016X/%d\t\n",
			b.GetValue(),
			b.GetLength(),
		); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func writeContactInfo(out io.Writer, name string, contactInfo *packetbroker.ContactInfo) error {
	if contactInfo == nil {
		return nil
	}
	return WriteKV(out,
		fmt.Sprintf("%s Name", name), contactInfo.GetName(),
		fmt.Sprintf("%s Email", name), contactInfo.GetEmail(),
		fmt.Sprintf("%s URL", name), contactInfo.GetUrl(),
	)
}

func x509SubjectFromPair(certPEMBlock, keyPEMBlock []byte) (string, error) {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return "", fmt.Errorf("parse TLS client certificate and key: %w", err)
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("parse X.509 certificate: %w", err)
	}
	return x509Cert.Subject.String(), nil
}

// x509Subjects returns the subjects of the certificates in the PEM encoded data.
// Non-certificate blocks and certificates that cannot be parsed are skipped.
func x509Subjects(pemCerts []byte) []string {
	var subjects []string
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		subjects = append(subjects, cert.Subject.String())
	}
	return subjects
}

func writeTarget(out io.Writer, target *packetbroker.Target, verbose bool) error {
	if target == nil {
		return nil
	}

	targetAuthKV := func(auth *packetbroker.Target_Authentication) []string {
		switch a := auth.GetValue().(type) {
		case *packetbroker.Target_Authentication_PbTokenAuth:
			return []string{"Token", "Packet Broker"}
		case *packetbroker.Target_Authentication_BasicAuth:
			kv := []string{"Basic username", a.BasicAuth.GetUsername()}
			if verbose {
				kv = append(kv, "Basic password", a.BasicAuth.GetPassword())
			}
			return kv
		case *packetbroker.Target_Authentication_CustomAuth:
			return []string{"Authorization", a.CustomAuth.GetValue()}
		case *packetbroker.Target_Authentication_TlsClientAuth:
			if verbose {
				return []string{
					"TLS client certificate", string(a.TlsClientAuth.GetCert()),
					"TLS client key", string(a.TlsClientAuth.GetKey()),
				}
			}
			sub, err := x509SubjectFromPair(a.TlsClientAuth.GetCert(), a.TlsClientAuth.GetKey())
			if err != nil {
				return []string{"Error", err.Error()}
			}
			return []string{"TLS client certificate subject", sub}
		default:
			return nil
		}
	}

	var authKV []string
	switch a := target.GetDefaultAuthentication().(type) {
	case *packetbroker.Target_PbTokenAuth:
		authKV = targetAuthKV(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_PbTokenAuth{},
		})
	case *packetbroker.Target_BasicAuth_:
		authKV = targetAuthKV(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_BasicAuth{
				BasicAuth: a.BasicAuth,
			},
		})
	case *packetbroker.Target_CustomAuth_:
		authKV = targetAuthKV(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_CustomAuth{
				CustomAuth: a.CustomAuth,
			},
		})
	case *packetbroker.Target_TlsClientAuth:
		authKV = targetAuthKV(&packetbroker.Target_Authentication{
			Value: &packetbroker.Target_Authentication_TlsClientAuth{
				TlsClientAuth: a.TlsClientAuth,
			},
		})
	}

	var rootCAs []string
	if cas := target.GetRootCas(); len(cas) > 0 {
		rootCAs = x509Subjects(cas)
	}

	if err := WriteKV(out,
		"Target Protocol", target.GetProtocol().String(),
		"Target Address", target.GetAddress(),
		"Target fNS Path", target.GetFNsPath(),
		"Target sNS Path", target.GetSNsPath(),
		"Target hNS Path", target.GetHNsPath(),
	); err != nil {
		return err
	}
	for i, subject := range rootCAs {
		if err := WriteKV(out, fmt.Sprintf("Target Root CA #%d", i+1), subject); err != nil {
			return err
		}
	}
	if len(authKV) > 0 {
		if err := writeLine(out, "\nTarget Authorization"); err != nil {
			return err
		}
		if err := WriteKV(out, authKV...); err != nil {
			return err
		}
	}

	// TODO: Sort by NetID.
	for netID, auth := range target.GetOriginNetIdAuthentication() {
		if err := writeLine(out, fmt.Sprintf("\nTarget Authorization %s", packetbroker.NetID(netID))); err != nil {
			return err
		}
		if err := WriteKV(out, targetAuthKV(auth)...); err != nil {
			return err
		}
	}
	return nil
}

// WriteJoinServer writes the Join Server.
func WriteJoinServer(out io.Writer, joinServer *packetbroker.JoinServer, verbose bool) error {
	if err := WriteKV(out,
		"ID", fmt.Sprintf("%d", joinServer.GetId()),
		"Name", joinServer.GetName(),
	); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Administrator", joinServer.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Technical", joinServer.GetTechnicalContact()); err != nil {
		return err
	}
	switch resolver := joinServer.GetResolver().(type) {
	case *packetbroker.JoinServer_Fixed:
		if err := WriteKV(out,
			"Fixed NetID", packetbroker.NetID(resolver.Fixed.GetNetId()).String(),
			"Fixed Tenant ID", resolver.Fixed.GetTenantId(),
			"Fixed Cluster ID", resolver.Fixed.GetClusterId(),
		); err != nil {
			return err
		}
	case *packetbroker.JoinServer_Lookup:
		if err := writeTarget(out, resolver.Lookup, verbose); err != nil {
			return err
		}
	}
	if err := writeLine(out, "\nJoinEUI Prefixes:"); err != nil {
		return err
	}
	return WriteJoinEUIPrefixes(out, joinServer.GetJoinEuiPrefixes())
}

// WriteNetwork writes the Network.
func WriteNetwork(out io.Writer, network *packetbroker.Network, verbose bool) error {
	var (
		delegatedNetID   string
		representsNetIDs string
	)
	if val := network.GetDelegatedNetId(); val != nil {
		delegatedNetID = packetbroker.NetID(val.GetValue()).String()
	}
	if len(network.GetRepresentsNetIds()) > 0 {
		netIDs := make([]string, len(network.GetRepresentsNetIds()))
		for i, n := range network.GetRepresentsNetIds() {
			netIDs[i] = packetbroker.NetID(n).String()
		}
		representsNetIDs = strings.Join(netIDs, ", ")
	}
	if err := WriteKV(out,
		"NetID", packetbroker.NetID(network.GetNetId()).String(),
		"Authority", network.GetAuthority(),
		"Name", network.GetName(),
		"Delegated NetID", delegatedNetID,
		"Represents NetIDs", representsNetIDs,
	); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Administrator", network.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Technical", network.GetTechnicalContact()); err != nil {
		return err
	}
	if err := writeTarget(out, network.GetTarget(), verbose); err != nil {
		return err
	}
	if err := writeLine(out, "\nDevAddr Blocks:"); err != nil {
		return err
	}
	return WriteDevAddrBlocks(out, network.GetDevAddrBlocks())
}

// WriteTenant writes the Tenant.
func WriteTenant(out io.Writer, tenant *packetbroker.Tenant, verbose bool) error {
	if err := WriteKV(out,
		"NetID", packetbroker.NetID(tenant.GetNetId()).String(),
		"Tenant ID", tenant.GetTenantId(),
		"Authority", tenant.GetAuthority(),
		"Name", tenant.GetName(),
	); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Administrator", tenant.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeContactInfo(out, "Technical", tenant.GetAdministrativeContact()); err != nil {
		return err
	}
	if err := writeTarget(out, tenant.GetTarget(), verbose); err != nil {
		return err
	}
	if err := writeLine(out, "\nDevAddr Blocks:"); err != nil {
		return err
	}
	return WriteDevAddrBlocks(out, tenant.GetDevAddrBlocks())
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
func WritePolicies(out io.Writer, defaults bool, policies ...*packetbroker.RoutingPolicy) error {
	var buf strings.Builder
	buf.WriteString("Forwarder\t\t")
	if !defaults {
		buf.WriteString("Home Network\t\t")
	}
	buf.WriteString("J\tM\tA\tS\tL\tJ\tM\tA\t\n")

	for _, policy := range policies {
		fmt.Fprintf(&buf, "%s\t%s\t",
			packetbroker.NetID(policy.GetForwarderNetId()),
			policy.GetForwarderTenantId(),
		)
		if !defaults {
			netID, tenantID := policy.GetHomeNetworkNetId(), policy.GetHomeNetworkTenantId()
			if netID == 0 && tenantID == "" {
				buf.WriteString("\t\t")
			} else {
				fmt.Fprintf(&buf, "%s\t%s\t",
					packetbroker.NetID(policy.GetHomeNetworkNetId()),
					policy.GetHomeNetworkTenantId(),
				)
			}
		}
		for _, b := range []bool{
			policy.GetUplink().GetJoinRequest(),
			policy.GetUplink().GetMacData(),
			policy.GetUplink().GetApplicationData(),
			policy.GetUplink().GetLocalization(),
			policy.GetUplink().GetSignalQuality(),
		} {
			if b {
				buf.WriteString("▲")
			}
			buf.WriteString("\t")
		}
		for _, b := range []bool{
			policy.GetDownlink().GetJoinAccept(),
			policy.GetDownlink().GetMacData(),
			policy.GetDownlink().GetApplicationData(),
		} {
			if b {
				buf.WriteString("▼")
			}
			buf.WriteString("\t")
		}
		buf.WriteString("\n")
	}
	if _, err := io.WriteString(out, buf.String()); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

// WriteVisibilities writes the gateway visibilities as a table.
func WriteVisibilities(out io.Writer, defaults bool, visibilities ...*packetbroker.GatewayVisibility) error {
	var buf strings.Builder
	buf.WriteString("Forwarder\t\t")
	if !defaults {
		buf.WriteString("Home Network\t\t")
	}
	buf.WriteString("Lo\tAp\tAc\tFt\tCi\tSt\tFp\tPr\t\n")

	for _, v := range visibilities {
		fmt.Fprintf(&buf, "%s\t%s\t",
			packetbroker.NetID(v.GetForwarderNetId()),
			v.GetForwarderTenantId(),
		)
		if !defaults {
			fmt.Fprintf(&buf, "%s\t%s\t",
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
				buf.WriteString("x")
			}
			buf.WriteString("\t")
		}
		buf.WriteString("\n")
	}
	if _, err := io.WriteString(out, buf.String()); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
