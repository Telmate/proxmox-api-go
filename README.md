# Proxmox API Go


Proxmox API in golang. Work in progress.


## Build

```
go build -o proxmox-api-go
```


## Run


```
export PM_API_URL="https://xxxx.com:8006/api2/json"
export PM_USER=user@pam
export PM_PASS=password

./proxmox-api-go installQemu 123 proxmox-node-name < qemu1.json

./proxmox-api-go createQemu 123 proxmox-node-name < qemu1.json

./proxmox-api-go -debug start 123

./proxmox-api-go -debug stop 123


```


### Format

createQmu JSON Sample:
```
{
  "name": "golang1.test.com",
  "desc": "Test proxmox-api-go",
  "memory": 2048,
  "diskGB": 10,
  "storage": "local",
  "os": "l26",
  "cores": 2,
  "sockets": 1,
  "iso": "local:iso/ubuntu-14.04.5-server-amd64.iso",
  "nic": "virtio",
  "bridge": "vmbr0",
  "vlan": -1
}
```

### ISO requirements

Kickstart auto install

* partition /dev/vda
* network eth1

Network is temprorarily eth1 during the pre-provision phase.
