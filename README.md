# go-ntp-client

Very simple one query ntp client using [beevik/ntp](https://github.com/beevik/ntp)

## Examples

```
go-ntp-client time.cloudflare.com | jq
```

```
go-ntp-client -network udp4 time.cloudflare.com | jq
```
