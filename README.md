# Proxmox API Go

Proxmox API in golang. For /api2/json. Work in progress.

Starting with Proxmox 5.2 you can use cloud-init options.

## Build

```sh
go build -o proxmox-api-go
```

## Run

```sh
export PM_API_URL="https://xxxx.com:8006/api2/json"
export PM_USER=user@pam
export PM_PASS=password
export PM_OTP=otpcode (only if required)
export PM_HTTP_HEADERS=Key,Value,Key1,Value1 (only if required)

./proxmox-api-go installQemu proxmox-node-name < qemu1.json

./proxmox-api-go createQemu 123 proxmox-node-name < qemu1.json

./proxmox-api-go -debug start 123

./proxmox-api-go -debug stop 123

./proxmox-api-go cloneQemu template-name proxmox-node-name < clone1.json

./proxmox-api-go migrate 123 migrate-to-proxmox-node-name

./proxmox-api-go createQemuSnapshot vm-name snapshot_name

./proxmox-api-go deleteQemuSnapshot vm-name snapshot_name

./proxmox-api-go listQemuSnapshot vm-name

./proxmox-api-go rollbackQemu vm-name

./proxmox-api-go getResourceList

./proxmox-api-go getVmList

./proxmox-api-go getUserList

./proxmox-api-go getUser userid

./proxmox-api-go updateUserPassword userid password

./proxmox-api-go setUser userid password < user.json

./proxmox-api-go deleteUser userid

./proxmox-api-go getAcmeAccountList

./proxmox-api-go getAcmeAccount accountid

./proxmox-api-go createAcmeAccount accountid < acmeAccount.json

./proxmox-api-go updateAcmeAccountEmail accountid email0,email1,email2

./proxmox-api-go deleteAcmeAccount accountid

./proxmox-api-go getAcmePluginList

./proxmox-api-go getAcmePlugin pluginid

./proxmox-api-go setAcmePlugin pluginid < acmePlugin.json

./proxmox-api-go deleteAcmePlugin pluginid

./proxmox-api-go getMetricsServerList

./proxmox-api-go getMetricsServer metricsid

./proxmox-api-go setMetricsServer metricsid < metricsServer.json

./proxmox-api-go deleteMetricsServer metricsid

./proxmox-api-go getStorageList

./proxmox-api-go getStorage storageid

./proxmox-api-go createStorage storageid < storage.json

./proxmox-api-go updateStorage storageid < storage.json

./proxmox-api-go deleteStorage

./proxmox-api-go getNetworkList node

./proxmox-api-go getNetworkInterface node interfaceName

./proxmox-api-go createNetwork < network.json

./proxmox-api-go updateNetwork < network.json

./proxmox-api-go deleteNetwork node iface

./proxmox-api-go applyNetwork node

./proxmox-api-go revertNetwork node

./proxmox-api-go node reboot proxmox-node-name

./proxmox-api-go node shutdown proxmox-node-name

```

## Proxy server support

Just use the flag -proxy and specify your proxy url and port

```sh
./proxmox-api-go -proxy https://localhost:8080 start 123
```

### Format

createQemu JSON Sample:

```json
{
  "name": "golang1.test.com",
  "desc": "Test proxmox-api-go",
  "memory": 2048,
  "os": "l26",
  "cores": 2,
  "sockets": 1,
  "iso": "local:iso/ubuntu-14.04.5-server-amd64.iso",
  "disk": {
    "0": {
      "type": "virtio",
      "storage": "local",
      "storage_type": "dir",
      "size": "30G",
      "backup": true
    }
  },
  "efidisk": {
    "storage": "local",
    "pre-enrolled-keys": "1",
    "efitype": "4m"
  },
  "network": {
    "0": {
      "model": "virtio",
      "bridge": "nat"
    },
    "1": {
      "model": "virtio",
      "bridge": "vmbr0",
      "firwall": true,
      "backup": true,
      "tag": -1
    }
  },
  "usb": {
    "0": {
      "host": "0658:0200",
      "usb3": true
    }
  }
}
```

cloneQemu JSON Sample:

```json
{
  "name": "golang2.test.com",
  "desc": "Test proxmox-api-go clone",
  "storage": "local",
  "memory": 2048,
  "cores": 2,
  "sockets": 1,
  "fullclone": 1
}
```

cloneQemu cloud-init JSON Sample:

```json
{
  "name": "cloudinit.test.com",
  "desc": "Test proxmox-api-go clone",
  "storage": "local",
  "memory": 2048,
  "cores": 2,
  "sockets": 1,
  "ipconfig": {
    "0": "gw=10.0.2.2,ip=10.0.2.17/24"
  },
  "sshkeys": "...",
  "nameserver": "8.8.8.8"
}
```

setUser JSON Sample:

```json
{
  "comment": "",
  "email": "b.wayne@proxmox.com",
  "enable": true,
  "expire": 0,
  "firstname": "Bruce",
  "lastname": "Wayne",
  "groups": [
    "admins",
    "usergroup"
  ],
  "keys": "2fa key"
}
```

createAcmeAccount JSON Sample:

```json
{
  "contact": [
    "b.wayne@proxmox.com",
    "c.kent@proxmox.com"
  ],
  "directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
  "tos": true
}
```

setAcmePlugin JSON Sample:

```json
{
  "api": "aws",
  "data": "AWS_ACCESS_KEY_ID=DEMOACCESSKEYID\nAWS_SECRET_ACCESS_KEY=DEMOSECRETACCESSKEY\n",
  "enable": true,
  "validation-delay": 30
}
```

setMetricsServer JSON Sample:

```json
{
  "port": 8086,
  "server": "192.168.67.3",
  "type": "influxdb",
  "enable": true,
  "mtu": 1500,
  "timeout": 1,
  "influxdb": {
    "protocol": "https",
    "max-body-size": 25000000,
    "verify-certificate": false,
    "token": "Rm8mqheWSVrrKKBW"
  }
}
```

createStorage JSON Sample:

```json
{
  "enable": true,
  "type": "smb",
  "smb": {
    "username": "b.wayne",
    "share": "NetworkShare",
    "preallocation": "metadata",
    "domain": "organization.com",
    "server": "10.20.1.1",
    "version": "3.11",
    "password": "Enter123!"
  },
  "content": {
    "backup": true,
    "iso": false,
    "template": true,
    "diskimage": true,
    "container": true,
    "snippets": false
  },
  "backupretention": {
    "last": 10,
    "hourly": 4,
    "daily": 7,
    "monthly": 3,
    "weekly": 2,
    "yearly": 1
  }
}
```

### Cloud-init options

Cloud-init VMs must be cloned from a cloud-init ready template.
See: https://pve.proxmox.com/wiki/Cloud-Init_Support

- ciuser - User name to change ssh keys and password for instead of the image’s configured default user.
- cipassword - Password to assign the user.
- cicustom - Specify custom files to replace the automatically generated ones at start.
- searchdomain - Sets DNS search domains for a container.
- nameserver - Sets DNS server IP address for a container.
- sshkeys - public ssh keys, one per line
- ipconfig0 - [gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]
- ipconfig1 - optional, same as ipconfig0 format

### ISO requirements (non cloud-init)

Kickstart auto install

- partition /dev/vda
- network eth1
- sshd (with preshared key/password)

Network is temprorarily eth1 during the pre-provision phase.

## Test

You're going to need [vagrant](https://www.vagrantup.com/downloads) and [virtualbox](https://www.virtualbox.org/wiki/Downloads) to run the tests:

```sh
make test
```
