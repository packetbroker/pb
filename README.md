# Packet Broker Clients

Packet Broker Clients are command-line utilities for working with [Packet Broker](https://www.packetbroker.org).

---

## Getting Started

### Installation

Make sure you have [Go](https://golang.org/doc/install) installed in your environment.

```bash
$ go get go.packetbroker.org/pb/cmd/pbadmin
$ go get go.packetbroker.org/pb/cmd/pbpub
$ go get go.packetbroker.org/pb/cmd/pbsub
```

### Configuration

In order to use the Packet Broker Clients, you need a client certificate signed by Packet Broker CA. Please see [Configuration](./configs) for more information.

Make sure you have `cert.pem`, `key.pem` and [`ca.pem`](./configs/ca.pem) in your working directory.

### Configure Policies

As a Forwarder, you can configure a default routing policy for all Home Networks, and routing policies per Home Network with `pbadmin`. 

As Forwarder `000042` with ID `example`, to see the default routing policy:

```bash
$ pbadmin policy get --forwarder-net-id 000042 --forwarder-id example --defaults
```

To see the routing policy for Home Network `C00123`:

```bash
$ pbadmin policy get --forwarder-net-id 000042 --forwarder-id example --home-network-net-id C00123
```

To see the routing policy for you own network:

```bash
$ pbadmin policy get --forwarder-net-id 000042 --forwarder-id example --home-network-net-id 000042
```

You can set policies by specifying letters from the following table:

| Policy | Uplink | Downlink |
| --- | :---: | :---: |
| Join-request/accept | `J` | `J` |
| MAC commands | `M` | `M` |
| Application data | `A` | `A` |
| Signal Quality | `S` | |
| Localization | `L` | |
| Allow Downlink | `D` | |

To enable all exchange by default:

```bash
$ pbadmin policy set --forwarder-net-id 000042 --forwarder-id example --defaults \
    --set-uplink JMASLD --set-downlink --JMA
```

To enable only device activation and MAC commands in both directions with Home Network `C00123`:

```bash
$ pbadmin policy set --forwarder-net-id 000042 --forwarder-id example --home-network-net-id C00123 \
    --set-uplink JMD --set-downlink --JM
```

To remove the uplink and downlink policy, use `--unset-uplink` and `--unset-downlink` respectively.

### Publish and Subscribe Traffic

To subscribe to routed downlink traffic:

```bash
$ pbsub --forwarder-net-id 000042 --forwarder-id example --group debug
```

To subscribe to routed uplink traffic:

```bash
$ pbsub --home-network-net-id 000042 --group debug
```

You can also subscribe to routed uplink traffic from your own Forwarder NetID and optionally a specific ID:

```bash
$ pbsub --home-network-net-id 000042 --forwarder-net-id 000042 --forwarder-id example --group debug
```

>**Important**: When using `pbsub`, always specify a shared subscription group that is different from the group you use in production. Otherwise, traffic gets split to your production subscriptions and your testing subscriptions.

To publish an uplink message for testing, you can pipe a JSON file to `pbpub`, specifying a Forwarder:

```bash
$ cat uplink.json | pbpub --forwarder-net-id 000042 --forwarder-id example
```

To publish a downlink message for testing, you can pipe a JSON file to `pbpub`, specifying a Home Network:

```bash
$ cat downlink.json | pbpub --home-network-net-id
```

See the [Examples](./examples) for example JSON files.

## Legal

Packet Broker Clients are Apache 2.0 licensed.
