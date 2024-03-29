// Copyright © 2020 The Things Industries B.V.

package pbflag

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	packetbroker "go.packetbroker.org/api/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type netIDValue struct {
	*packetbroker.NetID
}

func (f *netIDValue) String() string {
	if f.NetID == nil {
		return ""
	}
	return packetbroker.NetID(*f.NetID).String()
}

func (f *netIDValue) Set(s string) error {
	var netID packetbroker.NetID
	if err := netID.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	*f = netIDValue{
		NetID: &netID,
	}
	return nil
}

func (f *netIDValue) Type() string {
	return "netID"
}

func actorf(actor, name string) string {
	if actor != "" {
		return fmt.Sprintf("%s-%s", actor, name)
	}
	return name
}

// NetID returns flags for a NetID.
func NetID(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(netIDValue), actorf(actor, "net-id"), "LoRa Alliance NetID")
	return flags
}

// GetNetID returns the NetID from the flags.
func GetNetID(flags *flag.FlagSet, actor string) (packetbroker.NetID, bool) {
	netID := flags.Lookup(actorf(actor, "net-id")).Value.(*netIDValue).NetID
	if netID == nil {
		return 0, false
	}
	return packetbroker.NetID(*netID), true
}

// TenantID returns flags for a TenantID.
func TenantID(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(NetID(actor))
	flags.String(actorf(actor, "tenant-id"), "", "tenant ID")
	return flags
}

// GetTenantID returns the TenantID from the flags.
// The actor is used as prefix.
func GetTenantID(flags *flag.FlagSet, actor string) (packetbroker.TenantID, bool) {
	netID, ok := GetNetID(flags, actor)
	if !ok {
		return packetbroker.TenantID{}, false
	}
	tenantID, _ := flags.GetString(actorf(actor, "tenant-id"))
	return packetbroker.TenantID{
		NetID: netID,
		ID:    tenantID,
	}, true
}

// GetTenantIDWrappers returns the TenantID as protobuf wrappers.
// The returned values are nil when they are not set.
func GetTenantIDWrappers(flags *flag.FlagSet, actor string) (netID *wrapperspb.UInt32Value, id *wrapperspb.StringValue) {
	tntID, ok := GetTenantID(flags, actor)
	if !ok {
		return
	}
	netID = wrapperspb.UInt32(uint32(tntID.NetID))
	if TenantIDChanged(flags, actor) {
		id = wrapperspb.String(tntID.ID)
	}
	return
}

// TenantIDChanged returns whether the tenant ID flag has been changed explicitly.
func TenantIDChanged(flags *flag.FlagSet, actor string) bool {
	return flags.Changed(actorf(actor, "tenant-id"))
}

// Endpoint returns flags for an Endpoint.
func Endpoint(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(TenantID(actor))
	flags.String(actorf(actor, "cluster-id"), "", "cluster ID")
	return flags
}

// GetEndpoint returns the Endpoint from the flags.
func GetEndpoint(flags *flag.FlagSet, actor string) (packetbroker.Endpoint, bool) {
	tenantID, ok := GetTenantID(flags, actor)
	if !ok {
		return packetbroker.Endpoint{}, false
	}
	clusterID, _ := flags.GetString(actorf(actor, "cluster-id"))
	return packetbroker.Endpoint{
		TenantID:  tenantID,
		ClusterID: clusterID,
	}, true
}

// HasEndpoint returns which endpoint flags are set.
func HasEndpoint(flags *flag.FlagSet, actor string) (hasNetID, hasTenantID, hasClusterID bool) {
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case actorf(actor, "net-id"):
			hasNetID = true
		case actorf(actor, "tenant-id"):
			hasTenantID = true
		case actorf(actor, "cluster-id"):
			hasClusterID = true
		}
	})
	return
}

// EndpointFlagsChanged returns whether any of the endpoint flags are changed.
func EndpointFlagsChanged(flags *flag.FlagSet, actor string) bool {
	return flags.Changed(actorf(actor, "net-id")) ||
		flags.Changed(actorf(actor, "tenant-id")) ||
		flags.Changed(actorf(actor, "cluster-id"))
}

// ContactInfo returns flags for contact information.
func ContactInfo(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.String(actorf(actor, "name"), "", "name")
	flags.String(actorf(actor, "email"), "", "email address")
	flags.String(actorf(actor, "url"), "", "url")
	return flags
}

// GetContactInfo returns the contact information from flags.
func GetContactInfo(flags *flag.FlagSet, actor string) *packetbroker.ContactInfo {
	name, _ := flags.GetString(actorf(actor, "name"))
	email, _ := flags.GetString(actorf(actor, "email"))
	url, _ := flags.GetString(actorf(actor, "url"))
	if name == "" && email == "" && url == "" {
		return nil
	}
	return &packetbroker.ContactInfo{
		Name:  name,
		Email: email,
		Url:   url,
	}
}

type devAddrBlocksValue []*packetbroker.DevAddrBlock

func (f *devAddrBlocksValue) String() string {
	ss := make([]string, len(*f))
	for i, b := range *f {
		prefix, _ := b.Prefix.MarshalText()
		if b.HomeNetworkClusterId != "" {
			ss[i] = fmt.Sprintf("%s=%s", string(prefix), b.HomeNetworkClusterId)
		} else {
			ss[i] = string(prefix)
		}
	}
	return strings.Join(ss, ",")
}

func (f *devAddrBlocksValue) Set(s string) error {
	if s == "" {
		*f = []*packetbroker.DevAddrBlock{}
		return nil
	}
	blocks := strings.Split(s, ",")
	res := make([]*packetbroker.DevAddrBlock, len(blocks))
	for i, b := range blocks {
		parts := strings.SplitN(b, "=", 2)
		res[i] = &packetbroker.DevAddrBlock{
			Prefix: &packetbroker.DevAddrPrefix{},
		}
		if len(parts) == 2 {
			res[i].HomeNetworkClusterId = parts[1]
		}
		if err := res[i].Prefix.UnmarshalText([]byte(parts[0])); err != nil {
			return err
		}
	}
	*f = res
	return nil
}

func (f *devAddrBlocksValue) Type() string {
	return "devAddrBlocks"
}

// DevAddrBlocks returns flags for DevAddr blocks.
func DevAddrBlocks(addRemove bool) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(devAddrBlocksValue), "dev-addr-blocks", "DevAddr blocks")
	if addRemove {
		flags.Var(new(devAddrBlocksValue), "dev-addr-blocks-add", "DevAddr blocks to add")
		flags.Var(new(devAddrBlocksValue), "dev-addr-blocks-remove", "DevAddr blocks to remove")
	}
	return flags
}

// GetDevAddrBlocks returns the DevAddr blocks from the flags.
func GetDevAddrBlocks(flags *flag.FlagSet) (all, add, remove []*packetbroker.DevAddrBlock) {
	all = []*packetbroker.DevAddrBlock(*flags.Lookup("dev-addr-blocks").Value.(*devAddrBlocksValue))
	if f := flags.Lookup("dev-addr-blocks-add"); f != nil {
		add = []*packetbroker.DevAddrBlock(*f.Value.(*devAddrBlocksValue))
	}
	if f := flags.Lookup("dev-addr-blocks-remove"); f != nil {
		remove = []*packetbroker.DevAddrBlock(*f.Value.(*devAddrBlocksValue))
	}
	return
}

type joinEUIPrefixesValue []*packetbroker.JoinEUIPrefix

func (f *joinEUIPrefixesValue) String() string {
	ss := make([]string, len(*f))
	for i, b := range *f {
		prefix, _ := b.MarshalText()
		ss[i] = string(prefix)
	}
	return strings.Join(ss, ",")
}

func (f *joinEUIPrefixesValue) Set(s string) error {
	if s == "" {
		*f = []*packetbroker.JoinEUIPrefix{}
		return nil
	}
	blocks := strings.Split(s, ",")
	res := make([]*packetbroker.JoinEUIPrefix, len(blocks))
	for i, b := range blocks {
		res[i] = new(packetbroker.JoinEUIPrefix)
		if err := res[i].UnmarshalText([]byte(b)); err != nil {
			return err
		}
	}
	*f = res
	return nil
}

func (f *joinEUIPrefixesValue) Type() string {
	return "joinEUIPrefixes"
}

// JoinEUIPrefixes returns flags for JoinEUI prefixes.
func JoinEUIPrefixes() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(joinEUIPrefixesValue), "join-eui-prefixes", "JoinEUI prefixes")
	return flags
}

// GetJoinEUIPrefixes returns the JoinEUI prefixes from the flags.
func GetJoinEUIPrefixes(flags *flag.FlagSet) []*packetbroker.JoinEUIPrefix {
	blocks := flags.Lookup("join-eui-prefixes").Value.(*joinEUIPrefixesValue)
	return []*packetbroker.JoinEUIPrefix(*blocks)
}

type monthYear struct {
	valid       bool
	month, year int
}

func (f *monthYear) String() string {
	return fmt.Sprintf("%04d-%02d", f.year, f.month)
}

func (f *monthYear) Set(s string) error {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return errors.New("pbflag: invalid month year: expect YYYY-MM")
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("pbflag: invalid year %q: %w", parts[0], err)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("pbflag: invalid month %q: %w", parts[1], err)
	}
	if month < 1 || month > 12 {
		return fmt.Errorf("pbflag: invalid month %d", month)
	}
	*f = monthYear{
		valid: true,
		month: month,
		year:  year,
	}
	return nil
}

func (f *monthYear) Type() string {
	return "monthYear"
}

// MonthYear returns flags for a month in a year.
func MonthYear(name string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(monthYear), name, "month in a year (YYYY-MM)")
	return flags
}

// GetMonthYear returns the month year from flags.
func GetMonthYear(flags *flag.FlagSet, name string) (month, year int, ok bool) {
	monthYear := flags.Lookup(name).Value.(*monthYear)
	if !monthYear.valid {
		return 0, 0, false
	}
	return monthYear.month, monthYear.year, true
}

type messageType int

const (
	message messageType = iota
	messageDeliveryState
)

func (c messageType) String() string {
	switch c {
	case message:
		return "message"
	case messageDeliveryState:
		return "message-delivery-state"
	default:
		return "unknown"
	}
}

func (c *messageType) Set(s string) error {
	switch s {
	case "message":
		*c = message
	case "message-delivery-state":
		*c = messageDeliveryState
	default:
		return fmt.Errorf("pbflag: invalid message type: %s", s)
	}
	return nil
}

func (c *messageType) Type() string {
	return "messageType"
}

// MessageType returns flags for the message type.
func MessageType() *flag.FlagSet {
	flags := new(flag.FlagSet)
	v := messageType(message)
	flags.Var(&v, "message-type", "message type (message, message-delivery-state)")
	return flags
}

// NewForwarderMessage returns a new message based on the specified type.
func NewForwarderMessage(flags *flag.FlagSet) proto.Message {
	switch *flags.Lookup("message-type").Value.(*messageType) {
	case message:
		return new(packetbroker.UplinkMessage)
	case messageDeliveryState:
		return new(packetbroker.DownlinkMessageDeliveryStateChange)
	default:
		return nil
	}
}

// NewHomeNetworkMessage returns a new message based on the specified type.
func NewHomeNetworkMessage(flags *flag.FlagSet) proto.Message {
	switch *flags.Lookup("message-type").Value.(*messageType) {
	case message:
		return new(packetbroker.DownlinkMessage)
	case messageDeliveryState:
		return new(packetbroker.UplinkMessageDeliveryStateChange)
	default:
		return nil
	}
}

type protocol struct {
	*packetbroker.Protocol
}

func (p protocol) String() string {
	if p.Protocol == nil {
		return ""
	}
	return packetbroker.Protocol_name[int32(*p.Protocol)]
}

func (p *protocol) Set(s string) error {
	if s == "" {
		*p = protocol{}
		return nil
	}
	i, ok := packetbroker.Protocol_value[s]
	if !ok {
		return fmt.Errorf("pbflag: invalid protocol: %s", s)
	}
	*p = protocol{
		Protocol: (*packetbroker.Protocol)(&i),
	}
	return nil
}

func (p *protocol) Type() string {
	return "protocol"
}

// Protocol returns flags for a protocol.
func Protocol(actor string) *flag.FlagSet {
	names := make([]string, 0, len(packetbroker.Protocol_value))
	for k := range packetbroker.Protocol_value {
		names = append(names, k)
	}
	sort.Strings(names)
	flags := new(flag.FlagSet)
	flags.Var(new(protocol), actorf(actor, "protocol"), fmt.Sprintf("protocol (%s)", strings.Join(names, ",")))
	return flags
}

// TargetProtocolChanged returns whether the flag has been changed.
func TargetProtocolChanged(flags *flag.FlagSet, actor string) bool {
	return flags.Changed(actorf(actor, "protocol"))
}

// GetTargetProtocol returns the protocol from the flags.
func GetTargetProtocol(flags *flag.FlagSet, actor string) *packetbroker.Protocol {
	return flags.Lookup(actorf(actor, "protocol")).Value.(*protocol).Protocol
}

type apiKeyRightsValue []packetbroker.Right

func (p apiKeyRightsValue) String() string {
	rights := make([]string, 0, len(p))
	for _, v := range p {
		rights = append(rights, v.String())
	}
	return strings.Join(rights, ",")
}

func (p *apiKeyRightsValue) Set(s string) error {
	if s == "" {
		*p = []packetbroker.Right{}
		return nil
	}
	rights := strings.Split(s, ",")
	res := make([]packetbroker.Right, len(rights))
	for i, r := range rights {
		v, ok := packetbroker.Right_value[r]
		if !ok {
			return fmt.Errorf("pbflag: invalid right: %s", r)
		}
		res[i] = packetbroker.Right(v)
	}
	*p = res
	return nil
}

func (p *apiKeyRightsValue) Type() string {
	return "apiKeyRightsValue"
}

// APIKeyRights returns flags for API key rights.
func APIKeyRights(defaultRights ...packetbroker.Right) *flag.FlagSet {
	flags := new(flag.FlagSet)
	names := make([]string, 0, len(packetbroker.Right_value))
	for k := range packetbroker.Right_value {
		names = append(names, k)
	}
	sort.Strings(names)
	value := apiKeyRightsValue(defaultRights)
	flags.Var(&value, "rights", fmt.Sprintf("API key rights (%s)", strings.Join(names, ",")))
	return flags
}

// GetAPIKeyRights returns the API key rights from the flags.
func GetAPIKeyRights(flags *flag.FlagSet) []packetbroker.Right {
	rights := flags.Lookup("rights").Value.(*apiKeyRightsValue)
	return []packetbroker.Right(*rights)
}

type uplinkPolicyValue packetbroker.RoutingPolicy_Uplink

func (p *uplinkPolicyValue) String() string {
	var res string
	if p.JoinRequest {
		res += "J"
	}
	if p.MacData {
		res += "M"
	}
	if p.ApplicationData {
		res += "A"
	}
	if p.SignalQuality {
		res += "S"
	}
	if p.Localization {
		res += "L"
	}
	return res
}

func (p *uplinkPolicyValue) Set(s string) error {
	*p = uplinkPolicyValue(packetbroker.RoutingPolicy_Uplink{
		JoinRequest:     strings.ContainsRune(s, 'J'),
		MacData:         strings.ContainsRune(s, 'M'),
		ApplicationData: strings.ContainsRune(s, 'A'),
		SignalQuality:   strings.ContainsRune(s, 'S'),
		Localization:    strings.ContainsRune(s, 'L'),
	})
	return nil
}

func (p *uplinkPolicyValue) Type() string {
	return "uplinkPolicy"
}

type downlinkPolicyValue packetbroker.RoutingPolicy_Downlink

func (p *downlinkPolicyValue) String() string {
	var res string
	if p.JoinAccept {
		res += "J"
	}
	if p.MacData {
		res += "M"
	}
	if p.ApplicationData {
		res += "A"
	}
	return res
}

func (p *downlinkPolicyValue) Set(s string) error {
	*p = downlinkPolicyValue(packetbroker.RoutingPolicy_Downlink{
		JoinAccept:      strings.ContainsRune(s, 'J'),
		MacData:         strings.ContainsRune(s, 'M'),
		ApplicationData: strings.ContainsRune(s, 'A'),
	})
	return nil
}

func (p *downlinkPolicyValue) Type() string {
	return "downlinkPolicy"
}

// RoutingPolicy returns flags for a routing policy.
func RoutingPolicy() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(uplinkPolicyValue), "set-uplink", "uplink policy, use letters J, M, A, S and L")
	flags.Var(new(downlinkPolicyValue), "set-downlink", "downlink policy, use letters J, M and A")
	return flags
}

// GetRoutingPolicy returns the routing policy from the flags.
func GetRoutingPolicy(flags *flag.FlagSet) (*packetbroker.RoutingPolicy_Uplink, *packetbroker.RoutingPolicy_Downlink) {
	return (*packetbroker.RoutingPolicy_Uplink)(flags.Lookup("set-uplink").Value.(*uplinkPolicyValue)),
		(*packetbroker.RoutingPolicy_Downlink)(flags.Lookup("set-downlink").Value.(*downlinkPolicyValue))
}

type gatewayVisibilityValue packetbroker.GatewayVisibility

func (v *gatewayVisibilityValue) String() string {
	var res string
	if v.Location {
		res += "Lo"
	}
	if v.AntennaPlacement {
		res += "Ap"
	}
	if v.AntennaCount {
		res += "Ac"
	}
	if v.FineTimestamps {
		res += "Ft"
	}
	if v.ContactInfo {
		res += "Ci"
	}
	if v.Status {
		res += "St"
	}
	if v.FrequencyPlan {
		res += "Fp"
	}
	if v.PacketRates {
		res += "Pr"
	}
	return res
}

func (v *gatewayVisibilityValue) Set(s string) error {
	*v = gatewayVisibilityValue(packetbroker.GatewayVisibility{
		Location:         strings.Contains(s, "Lo"),
		AntennaPlacement: strings.Contains(s, "Ap"),
		AntennaCount:     strings.Contains(s, "Ac"),
		FineTimestamps:   strings.Contains(s, "Ft"),
		ContactInfo:      strings.Contains(s, "Ci"),
		Status:           strings.Contains(s, "St"),
		FrequencyPlan:    strings.Contains(s, "Fp"),
		PacketRates:      strings.Contains(s, "Pr"),
	})
	return nil
}

func (v *gatewayVisibilityValue) Type() string {
	return "gatewayVisibilityValue"
}

// GatewayVisibility returns flags for gateway visibility.
func GatewayVisibility() *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.Var(new(gatewayVisibilityValue), "set", "gateway visibility, use symbols Lo, Ap, Ac, Ft, Ci, St, Fp, Pr")
	return flags
}

// GetGatewayVisibility returns the gateway visibility from the flags.
func GetGatewayVisibility(flags *flag.FlagSet) *packetbroker.GatewayVisibility {
	return (*packetbroker.GatewayVisibility)(flags.Lookup("set").Value.(*gatewayVisibilityValue))
}

type apiKeyState packetbroker.APIKeyState

func (p apiKeyState) String() string {
	return packetbroker.APIKeyState_name[int32(p)]
}

func (p *apiKeyState) Set(s string) error {
	i, ok := packetbroker.APIKeyState_value[s]
	if !ok {
		return fmt.Errorf("pbflag: invalid API key state: %s", s)
	}
	*p = apiKeyState(i)
	return nil
}

func (p *apiKeyState) Type() string {
	return "apiKeyState"
}

// APIKeyState returns flags for an API key state.
func APIKeyState(name string) *flag.FlagSet {
	names := make([]string, 0, len(packetbroker.APIKeyState_value))
	for k := range packetbroker.APIKeyState_value {
		names = append(names, k)
	}
	sort.Strings(names)
	flags := new(flag.FlagSet)
	flags.Var(new(apiKeyState), name, fmt.Sprintf("API key state (%s)", strings.Join(names, ",")))
	return flags
}

// GetAPIKeyState returns the API key state from the flags.
func GetAPIKeyState(flags *flag.FlagSet, name string) packetbroker.APIKeyState {
	return packetbroker.APIKeyState(*flags.Lookup(name).Value.(*apiKeyState))
}

// Target returns flags for a target.
func Target(actor string) *flag.FlagSet {
	flags := new(flag.FlagSet)
	flags.AddFlagSet(Protocol(actor))
	flags.String(actorf(actor, "address"), "", "address (e.g. URL with HTTP basic authentication)")
	flags.String(actorf(actor, "fns-path"), "", "path for Forwarding Network Server (fNS)")
	flags.String(actorf(actor, "sns-path"), "", "path for Serving Network Server (sNS)")
	flags.String(actorf(actor, "hns-path"), "", "path for Home Network Server (hNS)")
	flags.Bool(actorf(actor, "pb-token"), false, "use Packet Broker token")
	flags.String(actorf(actor, "authorization"), "", "custom authorization value (e.g. HTTP Authorization header value)")
	flags.String(actorf(actor, "root-cas-file"), "", "path to PEM encoded root CAs")
	flags.String(actorf(actor, "tls-cert-file"), "", "path to PEM encoded client certificate")
	flags.String(actorf(actor, "tls-key-file"), "", "path to PEM encoded private key")
	flags.AddFlagSet(NetID(actorf(actor, "origin")))
	return flags
}

// TargetFlagsChanged returns whether any of the target values has been changed.
func TargetFlagsChanged(flags *flag.FlagSet, actor string) bool {
	return TargetProtocolChanged(flags, actor) ||
		flags.Changed(actorf(actor, "address")) ||
		flags.Changed(actorf(actor, "fns-path")) ||
		flags.Changed(actorf(actor, "sns-path")) ||
		flags.Changed(actorf(actor, "hns-path")) ||
		flags.Changed(actorf(actor, "pb-token")) ||
		flags.Changed(actorf(actor, "authorization")) ||
		flags.Changed(actorf(actor, "root-cas-file")) ||
		flags.Changed(actorf(actor, "tls-cert-file")) ||
		flags.Changed(actorf(actor, "tls-key-file")) ||
		flags.Changed(actorf(actor, "net-id"))
}

// ApplyToTarget applies the values from the flags to the given target.
func ApplyToTarget(flags *flag.FlagSet, actor string, target **packetbroker.Target) error {
	protocol := GetTargetProtocol(flags, actor)
	if *target == nil {
		if protocol == nil {
			return nil
		}
		*target = new(packetbroker.Target)
	}
	if protocol != nil {
		(*target).Protocol = *protocol
	}

	switch (*target).Protocol {
	case packetbroker.Protocol_TS002_V1_0, packetbroker.Protocol_TS002_V1_1:
		var url *url.URL
		if address, err := flags.GetString(actorf(actor, "address")); err == nil && address != "" {
			url, err = url.Parse(address)
			if err != nil {
				return err
			}
		}

		pbToken, _ := flags.GetBool(actorf(actor, "pb-token"))
		authorization, _ := flags.GetString(actorf(actor, "authorization"))
		tlsCertFile, _ := flags.GetString(actorf(actor, "tls-cert-file"))
		tlsKeyFile, _ := flags.GetString(actorf(actor, "tls-key-file"))

		var authentication *packetbroker.Target_Authentication
		switch {
		// Packet Broker token authentication.
		case pbToken:
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_PbTokenAuth{
					PbTokenAuth: &packetbroker.Target_PacketBrokerTokenAuth{},
				},
			}
		// HTTP basic authentication.
		case url != nil && url.User != nil:
			password, _ := url.User.Password()
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_BasicAuth{
					BasicAuth: &packetbroker.Target_BasicAuth{
						Username: url.User.Username(),
						Password: password,
					},
				},
			}
			url.User = nil
		// Custom HTTP authorization value.
		case authorization != "":
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_CustomAuth{
					CustomAuth: &packetbroker.Target_CustomAuth{
						Value: authorization,
					},
				},
			}
		// TLS client authentication.
		case tlsCertFile != "" || tlsKeyFile != "":
			tlsCert, err := ioutil.ReadFile(tlsCertFile)
			if err != nil {
				return err
			}
			tlsKey, err := ioutil.ReadFile(tlsKeyFile)
			if err != nil {
				return err
			}
			authentication = &packetbroker.Target_Authentication{
				Value: &packetbroker.Target_Authentication_TlsClientAuth{
					TlsClientAuth: &packetbroker.Target_TLSClientAuth{
						Cert: tlsCert,
						Key:  tlsKey,
					},
				},
			}
		}

		if netID, ok := GetNetID(flags, actorf(actor, "origin")); ok {
			if (*target).OriginNetIdAuthentication == nil {
				(*target).OriginNetIdAuthentication = make(map[uint32]*packetbroker.Target_Authentication)
			}
			if authentication == nil {
				delete((*target).OriginNetIdAuthentication, uint32(netID))
			} else {
				(*target).OriginNetIdAuthentication[uint32(netID)] = authentication
			}
		} else if authentication != nil {
			switch auth := authentication.GetValue().(type) {
			case *packetbroker.Target_Authentication_PbTokenAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_PbTokenAuth{
					PbTokenAuth: auth.PbTokenAuth,
				}
			case *packetbroker.Target_Authentication_BasicAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_BasicAuth_{
					BasicAuth: auth.BasicAuth,
				}
			case *packetbroker.Target_Authentication_CustomAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_CustomAuth_{
					CustomAuth: auth.CustomAuth,
				}
			case *packetbroker.Target_Authentication_TlsClientAuth:
				(*target).DefaultAuthentication = &packetbroker.Target_TlsClientAuth{
					TlsClientAuth: auth.TlsClientAuth,
				}
			}
		}

		if url != nil {
			(*target).Address = url.String()
		}
		for _, p := range []struct {
			target *string
			flag   string
		}{
			{&(*target).FNsPath, actorf(actor, "fns-path")},
			{&(*target).SNsPath, actorf(actor, "sns-path")},
			{&(*target).HNsPath, actorf(actor, "hns-path")},
		} {
			if flags.Changed(p.flag) {
				*p.target, _ = flags.GetString(p.flag)
			}
		}

	default:
		return fmt.Errorf("invalid protocol: %s", protocol)
	}

	if flags.Changed(actorf(actor, "root-cas-file")) {
		rootCAsFile, _ := flags.GetString(actorf(actor, "root-cas-file"))
		if rootCAsFile != "" {
			var err error
			(*target).RootCas, err = ioutil.ReadFile(rootCAsFile)
			if err != nil {
				return fmt.Errorf("read root CAs file %q: %w", rootCAsFile, err)
			}
		} else {
			(*target).RootCas = nil
		}
	}

	return nil
}
