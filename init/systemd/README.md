install the service file:

```
cp graylog-proxy.service /lib/systemd/system/  && ln -s /lib/systemd/system/graylog-proxy.service /etc/systemd/system/graylog-proxy.service
```

enable the service

```
systemctl enable graylog-proxy.service
```

start the service

```
systemctl start graylog-proxy.service
```