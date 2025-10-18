# `create storage` or `createStorage` (legacy) JSON Sample

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
