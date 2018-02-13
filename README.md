## Graylog Proxy

This application can act as a proxy between a node that listen on a
connectionless protocol like unixgram or udp and forward all messages to
a GELF input on a graylog server (over a secure connection). So you can
have the advances of a UDP deliver transport in your application and a
tcp (with ssl and client authentication) for delivering the log messages
to a graylog server. This could mean that some messages get lost but
you application will not block or fail when a message could not be
delivered.


```text

 ************* SERVER *************             ***************** SERVER *************
 *                                *             *                                    *
 *  (APPLICATION)                 *             *   (GRAYLOG APPLICATION)            *
 *       |                        *             *       |                            *
 *       `-[UDP]---> (NODE)       *             *       |                            *
 *                     |          *             *      /                             *
 *                     `--------------[TCP+SSL]-------                               *
 *********************************               *************************************


```

The node used for forwarding traffic to a graylog server support gzip,
zlib, uncompressed and chuncked packages and will convert those to a
normal tcp package that a the graylog server support.

Supported remote inputs are tcp, tcp+tls, http, https and  it also
support the use of a client authentication and self generated CA
certificates. There are commands provided to generate certificates that
can used for setting up de input and which the client can use
authenticate and secure the log transmission.

## Example

We will demonstrate a simple config/setup between a symfony application
and an graylog backend.

For this demonstration we use [gelf-php](https://packagist.org/packages/graylog2/gelf-php) in our application that we use as a monolog handler.

you can install this with composer for you project:

```shell
    composer require graylog2/gelf-php:^1.5
```

and in you service.xml you should create a couple services:

```xml

    <service id="gelf.publisher" class="Gelf\Publisher">
        <argument type="service" id="gelf.transporter"/>
    </service>

    <service id="gelf.transporter" class="Gelf\Transport\IgnoreErrorTransportWrapper">
        <argument type="service" id="gelf.udp_transporter"/>
    </service>

    <service id="gelf.udp_transporter" class="Gelf\Transport\UdpTransport">
        <argument>127.0.0.1</argument>
        <argument>12201</argument>
        <argument type="constant">Gelf\Transport\UdpTransport::CHUNK_SIZE_LAN</argument>
    </service>

    <service id="monolog.gelf_handler" class="Monolog\Handler\GelfHandler">
        <argument type="service" id="gelf.publisher"/>
    </service>

```

now in your monolog config you can add a new handler:

```
monolog:

    handlers:

        gelf:
            type: service
            id: zicht_monolog.gelf_handler
            level: allert
```

with this all `alert` records will be handled by the glef handler and send
to the graylog node.


Generate some certificates so we could setup a "GELF TCP" and using tls
and force client authentication:

```
mkdir /tmp/certs
# create CA root certificates
graylog-proxy --cwd=/tmp/certs create:ca 'CN=GrayLog Test CA'
# creat server key and certificate
graylog-proxy --cwd=/tmp/certs create:server 'CN=GrayLog Test Server'
# create client key and certificate for host example.com
graylog-proxy --cwd=/tmp/certs create:client example.com 'CN=GrayLog Test Client'
```

Now we can create a new input on the graylog server, use the Server.crt
for "TLS cert file", Server.pem for "TLS private key file", CA_Root.crt
as the "TLS Client Auth Trusted Certs" and set the
"TLS client authentication" on required.

On the "example.com" server we have to copy Client.crt, Client.pem and
CA_Root.crt to /etc/graylog-proxy/cert and start the node:

```
graylog-proxy --cwd=/etc/graylog-proxy/cert listen udp://127.0.0.1:12201 tcp+ssl://example.logger.com:12201
```