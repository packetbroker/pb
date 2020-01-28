# Packet Broker Client Configuration

Packet Broker uses TLS mutual authentication with Forwarders and Home Networks. In order to connect as Forwarder or Home Network, you need a TLS client certificate that is signed by the Packet Broker CA.

## Client Certificate

To obtain a TLS client certificate to connect to Packet Broker, you generate a certificate signing request (CSR).

### Generate Private Key

Generate a private key:

```bash
$ openssl genrsa -out key.pem 2048
```

You may use a different algorithm and key length.

### Configure OpenSSL

1. Open [`openssl.cnf`](./openssl.cnf)
2. Replace `subjectAltName` with the Packet Broker URI according the specified format

### Create Certificate Signing Request

```bash
$ openssl req -config openssl.cnf -new -key key.pem -out csr.pem
```

>Note: **Organization Name** must match the full LoRa Alliance member name, the **Common Name** may be a brand name, and **Email Address** must match the technical contact person of the LoRa Alliance member.

### Request Certificate

Please send the CSR file to [join@packetbroker.org](mailto:join@packetbroker.org).

## Configure Certificates

The `pbadmin`, `pbpub` and `pbsub` commands attempt to find the client certificate file `cert.pem` and key file `key.pem` in the working directory. You can configure paths to these files by using the flags `--cert-file` and `--key-file` respectively.

The commands attempt to find the server certificate `ca.pem` in the working directory. You can [download the Packet Broker CA](./ca.pem) and save it as `ca.pem` or pass the path to this file with the flag `--ca-file`.
