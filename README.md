# Proxmox API Go

Proxmox API in golang. For /api2/json. Work in progress.

Starting with Proxmox 5.2 you can use cloud-init options.

## Contributing

Want to help with the project please refer to our [style-guide](docs/style-guide/style-guide.md).

## Build

An example of usage of the SDK is the CLI in this project, it can be compiled using the following command :

```sh
make build
```

> [!NOTE]
> It is possible to generate autocompletion for the CLI by executing
> ```sh
> make bash_completion  # or install_bash_completion to install completion locally
> ```
> However, the autocompletion will only be usable if `proxmox-api-go` is in `$PATH` and `NEW_CLI` is set to `true` (see below the new CLI usage)

## Using the CLI

Start with configuring the CLI, you can do so by creating a local `.env` file in the root directory of the project and add the following environment variables:

```sh
PM_API_URL="https://xxxx.com:8006/api2/json"
PM_USER=user@pam
PM_PASS=password
PM_OTP=otpcode (only if required)
PM_HTTP_HEADERS=Key,Value,Key1,Value1 (only if required)
```

> [!WARNING]
> Do not commit your local `.env` file to version control to keep your credentials secure.

It is also possible to export the environment variables:

```sh
export PM_API_URL="https://xxxx.com:8006/api2/json"
export PM_USER=user@pam
export PM_PASS=password
export PM_OTP=otpcode (only if required)
export PM_HTTP_HEADERS=Key,Value,Key1,Value1 (only if required)
```

### The new CLI

In order to use the new CLI, the environment variable `NEW_CLI` must be equal to `true` like that :

```sh
export NEW_CLI=true
# unset NEW_CLI  # To revert this operation
```

Then, it is possible to use `./proxmox-api-go help` to browse available commands.

### The legacy CLI

This is the default mode, and it is **deprecated**, however here is a list of commands (examples, not a complete list):

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

With the new CLI :

```sh
./proxmox-api-go --proxy https://localhost:8080 <cmd>
```

With the legacy CLI :

```sh
./proxmox-api-go -proxy https://localhost:8080 start 123
```

## JSON Config format

Many commands use configs from files or `stdin`, examples of these configs can be found in [this directory](./docs/config-examples).

## ISO requirements (non cloud-init)

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
