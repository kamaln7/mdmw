<p align="center">
  <img src="/mdmw.png" alt="mdmw logo" width="301" />
</p>

mdmw is a **M**ark**d**own **m**iddle**w**are HTTP server. It renders Markdown files and serves them back as HTML.

mdmw supports two storage drivers:

- **filesystem** uses the OS filesystem to look up files
- **spaces** uses [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) to look up files

### installation

1. download a binary suitable for your OS from [the releases page](https://github.com/kamaln7/mdmw/releases)
2. place said binary in `/usr/local/bin` or wherever you would like

### usage

1. by default, mdmw listens on `localhost:4000`. If you’d like to change that, use the `—listen_address` or `LISTEN_ADDRESS` flag.
2. WIP

### license

see [LICENSE](/LICENSE)