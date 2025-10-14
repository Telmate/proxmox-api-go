# JSON Config examples for commands

Many configs are passed through `stdin` or by specifying a file (with the parameter `--file` with the new CLI), here are examples about them :

* [`create guest qemu` or `createQemu` (legacy)](./clone-qemu.md)
* [`create acmeaccount` or `createAcmeAccount` (legacy)](./create-acme-account.md)
* [`create guest qemu` or `createQemu` (legacy)](./create-qemu.md)
* [`create storage` or `createStorage` (legacy)](./create-storage.md)
* [`setAcmePlugin` (legacy)](./set-acme-plugin.md)
* [`set metricserver` or `setMetricsServer` (legacy)](./set-metrics-server.md)
* [`set user` or `setUser` (legacy)](./set-user.md)

## Cloud init

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
