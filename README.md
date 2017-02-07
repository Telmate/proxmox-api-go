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

./proxmox-api-go -debug start 123

./proxmox-api-go -debug stop 123
```