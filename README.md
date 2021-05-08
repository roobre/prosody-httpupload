# prosody-httpupload

An HTTP upload server for Prosody's [mod_http_upload_external](https://modules.prosody.im/mod_http_upload_external.html)
, with focus in strictness, security, and ease of deployment in cloud environments.

## Motivation

This is similar to existing implementations, and particularly
to [prosody filer](https://github.com/ThomasLeister/prosody-filer), which is also written in Go. However I wanted
something more suited to my personal tastes, and thus this project was born.

## Usage

This service accepts the following configuration options:

| Name | Description | Default|
|------|-------------|--------|
| `listen-address` | Address where the server will listen on | `:8889` |
| `storage-path` | Path to uploaded files. RW access is needed in this folder | `data`, relative to CWD |
| `secret` | Secret string, used to authenticate clients. Must be set in accordance with Prosody's. | (empty, must be set by the user) |

Arguments can be supplied either through the command line as flags, or through the environment:

| Flag | Environment |
|------|-------------|
| `-listen-address` |  `HTTPUP_LISTEN_ADDRESS` |
| `-storage-path` | `HTTPUP_STORAGE_PATH` |
| `-secret` | `HTTPUP_SECRET` |

## FAQ

### Does this support a config file?

> No, it does not. Only env vars and flags are supported.

### Is there a Docker image available?

> Yes! Check out the `Dockerfile` or https://hub.docker.com/u/roobre/prosody-httpupload.

### Is there a Helm Chart available?

> Not yet. However, deployment is simple enough to fit in just a `Deployment`, `Service` and optionally a `Secret` object.
> A sample manifest is provided in the repo root.

### How do I configure HTTPs?

> It is not possible to configure TLS directly in this service. I strongly suggest using a robust reverse proxy such as
nginx or Caddy for this task.

### How can I make this server accept requests in a different path, e.g. `/upload/service`?

> You can achieve this by defining this server as an upstream for the desired path in your favorite reverse proxy:

```nginx
location /upload/service {
    proxy_pass http://prosody-httpupload:8889/; # Trailing slash strips path upstream
}
```

```Caddyfile
route /upload/service/* {
	uri strip_prefix /upload/service
	reverse_proxy http://prosody-httpupload:8889
}
```
