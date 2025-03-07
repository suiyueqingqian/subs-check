# Configuration File Details

### check

```yaml
check:
  items:
    - speed
    - youtube
    - openai
    - netflix
    - disney
  concurrent: 10
  timeout: 10
  interval: 10
```

- `items`: Check items, available options: `speed`, `youtube`, `openai`, `netflix`, `disney`
- `concurrent`: Number of concurrent checks
- `timeout`: Timeout duration in milliseconds
- `interval`: Check interval in minutes

### save

```yaml
save:
  method: http
  port: 8081
  webdav-url: "https://webdav-url/dav/"
  webdav-username: "webdav-username"
  webdav-password: "webdav-password"
  github-token: "github-token"
  github-gist-id: "github-gist-id"
  github-api-mirror: "https://worker-url/github"
  worker-url: https://worker-url
  worker-token: token 
```

- `method`: Save method, available options: `webdav`, `http`, `gist`, `r2`
- `port`: Save port
- webdav:
  - `webdav-url`: WebDAV URL
  - `webdav-username`: WebDAV username
  - `webdav-password`: WebDAV password
- gist:
  - `github-token`: GitHub token for Gist
  - `github-gist-id`: Gist ID
- r2:
  - `worker-url`: Worker URL
  - `worker-token`: Worker token

## mihomo

```yaml
# mihomo api
mihomo-api-url: "http://192.168.31.11:9090"
# mihomo api secret
mihomo-api-secret: ""
```

- `api-url`: Mihomo API URL
- `api-secret`: Mihomo API secret

## rename

```yaml
rename:
  flag: true
  method: "mix"
```

- `flag`: Whether to enable renaming
- `method`: Renaming method, available options: `mix`, `api`, `regex`

> When using the `mix` method, it will first perform `regex` renaming followed by `api` renaming 