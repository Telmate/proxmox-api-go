# Proxmox API Go

Proxmox API in golang. For /api2/json. Work in progress.

Starting with Proxmox 5.2 you can use cloud-init options.

## Contributing

Want to help with the project please refer to our [style-guide](docs/style-guide/style-guide.md).

## Build

```sh
go build -o proxmox-api-go
```

## Run

Create a local `.env` file in the root directory of the project and add the following environment variables:

```sh
PM_API_URL="https://xxxx.com:8006/api2/json"
PM_USER=user@pam
PM_PASS=password
PM_OTP=otpcode (only if required)
PM_HTTP_HEADERS=Key,Value,Key1,Value1 (only if required)
```

**Note**: Do not commit your local `.env` file to version control to keep your credentials secure.

Or export the environment variables:

```sh
export PM_API_URL="https://xxxx.com:8006/api2/json"
export PM_USER=user@pam
export PM_PASS=password
export PM_OTP=otpcode (only if required)
export PM_HTTP_HEADERS=Key,Value,Key1,Value1 (only if required)
```

Run commands (examples, not a complete list):

```sh
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

./proxmox-api-go unlink 123 proxmox-node-name "virtio1,virtio2,virtioN" [false|true]

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
  "description": "Test proxmox-api-go",
  "memory": {
    "capacity": 2048
  },
  "ostype": "l26",
  "cores": 2,
  "sockets": 1,
  "iso": {
    "storage": "local",
    "file": "ubuntu-14.04.5-server-amd64.iso"
  },
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
      "firewall": true,
      "tag": -1
    }
  },
  "rng0": {
    "source": "/dev/urandom",
    "max_bytes": "1024",
    "period": "1000"
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
  "description": "Test proxmox-api-go clone",
  "storage": "local",
  "memory": {
    "capacity": 2048
  },
  "cores": 2,
  "sockets": 1,
  "fullclone": 1
}
```

cloneQemu cloud-init JSON Sample:

```json
{
  "name": "cloudinit.test.com",
  "description": "Test proxmox-api-go clone",
  "storage": "local",
  "memory": {
    "capacity": 2048
  },
  "cores": 2,
  "sockets": 1,
  "cloudinit": {
    "ipconfig": {
      "0": {
        "ip4": {
          "address": "10.0.2.17/24",
          "gateway": "10.0.2.2"
        }
      }
    },
    "sshkeys": [
      "..."
    ],
    "dns": {
      "nameservers": [
        "8.8.8.8"
      ]
    }
  }
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

- `username` - User name to change ssh keys and password for instead of the image’s configured default user.
- `userpassword` - Password to assign the user.
- `cicustom` - Specify custom files to replace the automatically generated ones at start, as JSON object of:
  - `meta` - JSON object of:
    - `path` - as string
    - `storage` - as string
  - `network` - JSON object of:
    - `path` - as string
    - `storage` - as string
  - `user` - JSON object of:
    - `path` - as string
    - `storage` - as string
  - `vendor` - JSON object of:
    - `path` - as string
    - `storage` - as string
- `dns` - Sets DNS settings, as JSON object of:
  - `nameservers` - Sets DNS server IP address for a VM, as list of stings.
  - `searchdomain` - Sets DNS search domains for a VM, as string.
- `sshkeys` - public ssh keys, as list of strings.
- `ipconfig` - Sets IP configuration for network interfaces, as dictionary of interface index number to JSON object:
  - `ip4` - for IPv4, as JSON object of:
    - `address` - IPv4Format/CIDR
    - `dhcp` - true/false
    - `gateway` - GatewayIPv4
  - `ip6` - for IPv6, as JSON object of:
    - `address` - IPv6Format/CIDR
    - `dhcp` - true/false
    - `gateway` - GatewayIPv6
    - `slaac` - true/false
- `ciupgrade` - If true does a package update after startup

Example:

```json
{
  "cloudinit": {
    "username": "test",
    "userpassword": "passw0rd",
    "cicustom": {
      "meta": {
        "path": "path/to/meta",
        "storage": "local"
      },
      "network": {
        "path": "path/to/network",
        "storage": "local"
      },
      "user": {
        "path": "path/to/user",
        "storage": "local"
      },
      "vendor": {
        "path": "path/to/vendor",
        "storage": "local"
      }
    },
    "dns": {
      "nameservers": [
        "8.8.8.8"
      ],
      "searchdomain": "test.com"
    },
    "sshkeys": [
      "...",
      "..."
    ],
    "ipconfig": {
      "0": {
        "ip4": {
          "address": "10.0.2.17/24",
          "gateway": "10.0.2.2"
        },
        "ip6": {
          "address": "2001:0db8:0:2::17/64",
          "gateway": "2001:0db8:0:2::2"
        }
      },
      "1": {
        "ip4": {
          "dhcp": true
        },
        "ip6": {
          "dhcp": true
        }
      },
      "2": {
        "ip6": {
          "slaac": true
        }
      }
    },
    "ciupgrade": true
  }
}
```

### ISO requirements (non cloud-init)

Kickstart auto install

- partition /dev/vda
- network eth1
- sshd (with preshared key/password)

Network is temporarily eth1 during the pre-provision phase.

## Test

You're going to need [vagrant](https://www.vagrantup.com/downloads) and [virtualbox](https://www.virtualbox.org/wiki/Downloads) to run the tests:

```sh
make test
```
