# Webfront+

Webfront+ is an HTTP server and reverse proxy.

example :

```bash
webfront -apps=./apps.json -letsencrypt_cache=./letsencrypt_cache -hostname=${HOSTNAME}
```
and go to HOSTNAME to configure apps (set the given token to do so)

## Development

# Update dependencies

```bash
go get -u
go mod tidy
```
