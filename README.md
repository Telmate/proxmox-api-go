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

./proxmox-api-go installQemu proxmox-node-name < qemu1.json

./proxmox-api-go createQemu 123 proxmox-node-name < qemu1.json

./proxmox-api-go -debug start 123

./proxmox-api-go -debug stop 123

./proxmox-api-go cloneQemu template-name proxmox-node-name < clone1.json

./proxmox-api-go migrate pve1 123

./proxmox-api-go createQemuSnapshot vm-name snapshot_name

./proxmox-api-go deleteQemuSnapshot vm-name snapshot_name

./proxmox-api-go listQemuSnapshot vm-name

./proxmox-api-go rollbackQemu vm-name
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
  "ipconfig0": "gw=10.0.2.2,ip=10.0.2.17/24",
  "sshkeys" : "...",
  "nameserver": "8.8.8.8"
}
```

### Cloud-init options

Cloud-init VMs must be cloned from a cloud-init ready template. 
See: https://pve.proxmox.com/wiki/Cloud-Init_Support

* ciuser - User name to change ssh keys and password for instead of the image’s configured default user.
* cipassword - Password to assign the user. 
* cicustom - Specify custom files to replace the automatically generated ones at start.
* searchdomain - Sets DNS search domains for a container.
* nameserver - Sets DNS server IP address for a container.
* sshkeys - public ssh keys, one per line
* ipconfig0 - [gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]
* ipconfig1 - optional, same as ipconfig0 format

### ISO requirements (non cloud-init)

Kickstart auto install

* partition /dev/vda
* network eth1
* sshd (with preshared key/password)

Network is temprorarily eth1 during the pre-provision phase.

## Test

You're going to need [vagrant](https://www.vagrantup.com/downloads) and [virtualbox](https://www.virtualbox.org/wiki/Downloads) to run the tests:

```sh
make test
```
