# JSON Config examples for CLI commands

Many configs are passed through `stdin` or by specifying a file (with the parameter `--file` with the new CLI), here are examples about them :

## `create guest qemu` or `createQemu` (legacy) JSON Sample:

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

## `cloneQemu ` (legacy) JSON Sample:

```json
{
  "full": {
    "node": "pve-latest",
    "name": "golang2.test.com",
    "pool": "my-pool",
    "storage": "local-zfs",
    "format": "raw"
  }
}
```

## `set user` or `setUser` (legacy) JSON Sample:

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

## `create acmeaccount` or `createAcmeAccount` (legacy) JSON Sample:

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

## `setAcmePlugin` (legacy) JSON Sample:

```json
{
  "api": "aws",
  "data": "AWS_ACCESS_KEY_ID=DEMOACCESSKEYID\nAWS_SECRET_ACCESS_KEY=DEMOSECRETACCESSKEY\n",
  "enable": true,
  "validation-delay": 30
}
```

## `set metricserver` or `setMetricsServer` (legacy) JSON Sample:

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

## `create storage` or `createStorage` (legacy) JSON Sample:

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

## Cloud-init options

Cloud-init VMs must be cloned from a cloud-init ready template.
See: https://pve.proxmox.com/wiki/Cloud-Init_Support

The cloud init format should follow this schema :

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
