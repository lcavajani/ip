# IP


## Remote IP

The root path `/` returns the IP of the client, this is similar
to what `http://checkip.amazonaws.com/` provides.

**Example:**

```bash
curl -s 127.0.0.1:8000/
127.0.0.1
```


## Headers

The path `/info` returns info about the HTTP request.

**Example:**

```
curl -s 127.0.0.1:8000/info | jq '.'
{
  "headers": {
    "Accept": [
      "*/*"
    ],
    "User-Agent": [
      "curl/7.60.0"
    ]
  },
  "host": "127.0.0.1:8000",
  "remoteAddr": "127.0.0.1:49962"
}
```


## IPCalc

The path `/ip4calcc` takes two parameters:

* `ip`
* `cidr`

**Examples:**

```bash
curl -s 127.0.0.1:8000/ip4calcc?ip=192.168.0.100\&cidr=24 | jq '.'
{
  "address": "192.168.0.100",
  "cidr": "24",
  "netmask": "255.255.255.0",
  "networkCidr": "192.168.0.0/24",
  "hostMin": "192.168.0.1",
  "hostMax": "192.168.0.254",
  "broadcast": "192.168.0.255",
  "hostsAvailable": "254",
  "hostsTotal": "256"
}
```

```bash
curl -s 127.0.0.1:8000/ip4calcc?ip=192.168.0.100\&cidr=30 | jq '.'
{
  "address": "192.168.0.100",
  "cidr": "30",
  "netmask": "255.255.255.252",
  "networkCidr": "192.168.0.100/30",
  "hostMin": "192.168.0.101",
  "hostMax": "192.168.0.102",
  "broadcast": "192.168.0.103",
  "hostsAvailable": "2",
  "hostsTotal": "4"
}
```

```bash
curl -s 127.0.0.1:8000/ip4calcc?ip=192.168.0.100\&cidr=31 | jq '.'
{
  "address": "192.168.0.100",
  "cidr": "31",
  "netmask": "255.255.255.254",
  "networkCidr": "192.168.0.100/31",
  "network": "192.168.0.100",
  "hostMin": "192.168.0.100",
  "hostMax": "192.168.0.101",
  "hostsAvailable": "2",
  "hostsTotal": "2"
}
```

```bash
curl -s 127.0.0.1:8000/ip4calcc?ip=192.168.0.100\&cidr=32 | jq '.'
{
  "address": "192.168.0.100",
  "cidr": "32",
  "netmask": "255.255.255.255",
  "hostMin": "192.168.0.100",
  "hostMax": "192.168.0.100",
  "hostsAvailable": "1",
  "hostsTotal": "1"
}
```
