<p align="center">
  <img src="/mdmw.png" alt="mdmw logo" width="301" />
</p>

mdmw is a **M**ark**d**own **m**iddle**w**are HTTP server. It renders Markdown files and serves them back as HTML. It's not meant to be your sole www-facing webserver as it cannot serve non-markdown files.

You can use mdmw to host a micro-blog, quickly share documents, host documentation, etc.

mdmw supports two storage drivers (i.e. where it pulls markdown files from):

- **filesystem** uses the OS filesystem to look up files
- **spaces** uses [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) to look up files

mdmw exposes an HTTP server and uses the URI as the path for the markdown files.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [installation](#installation)
  - [use the docker image RECOMMENDED](#use-the-docker-image-recommended)
  - [use a pre-built mdmw binary](#use-a-pre-built-mdmw-binary)
- [usage](#usage)
  - [options](#options)
  - [configuration examples](#configuration-examples)
    - [as a yaml config file](#as-a-yaml-config-file)
    - [as cli flags](#as-cli-flags)
    - [as environment variables](#as-environment-variables)
- [license](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

### installation

#### use the docker image RECOMMENDED

1. pull the latest docker image

```
docker pull kamaln7/mdmw:latest
```

#### use a pre-built mdmw binary

1. download a binary suitable for your OS from [the releases page](https://github.com/kamaln7/mdmw/releases)
2. place said binary in `/usr/local/bin` or wherever you would like

### usage

mdmw listens on `localhost:4000` by default. Refer to the options section below for details on how to change that. If you are using the docker image, you will need to expose port 4000
    
See the options section below on how to configure mdmw. You can run `docker run -p 4000 kamaln7/mdmw` (use `-e` or `--env-file` to pass configuration options) or run `mdmw` directly if you chose to not use Docker.

#### options

```
Usage:
  mdmw [flags]

Flags:
      --config string               config file (default is ./.mdmw.yaml)
      --filesystem.path string      path to markdown files (default "./files")
  -h, --help                        help for mdmw
      --listenaddress string        address to listen on (default "localhost:4000")
      --outputtemplate string       path to HTML output template
      --spaces.auth.access string   DigitalOcean Spaces access key
      --spaces.auth.secret string   DigitalOcean Spaces secret key
      --spaces.cache string         DigitalOcean Spaces cache time (default "0")
      --spaces.path string          DigitalOcean Spaces files path (default "/")
      --spaces.region string        DigitalOcean Spaces region
      --spaces.space string         DigitalOcean Spaces space name
      --storage string              storage driver to use (default "filesystem")
      --validateextension           validate that files have a markdown extension (default true)
```

There are three ways to configure mdmw. examples for each can be found below

1. use a config file (yaml, toml, json, etc.) and pass `-config ./path/to/config`
    * options become nested objects, see the example below
2. pass cli flags as described above
3. use environment variables
    * options become uppercase with periods replaced by underscores, see the example below

#### configuration examples

* listen on `0.0.0.0:8080`
* serve files from a Space in AMS3

##### as a yaml config file

```
listenaddress: 0.0.0.0:8080
storage: spaces
spaces:
  auth:
    access: ACCESS KEY GOES HERE
    secret: SECRET KEY GOES HERE
  region: ams3
  space: SPACE NAME GOES HERE
```

##### as cli flags

```
mdmw \
  --listenaddress 0.0.0.0:8000 \
  --storage spaces \
  --spaces.auth.access "ACCESS KEY GOES HERE" \
  --spaces.auth.secret "SECRET KEY GOES HERE" \
  --spaces.region ams3 \
  --spaces.space "SPACE NAME GOES HERE"
```

##### as environment variables

```
LISTENADDRESS=0.0.0.0:8000 \
STORAGE=spaces \
SPACES_AUTH_ACCESS=ACCESS KEY GOES HERE \
SPACES_AUTH_SECRET=SECRET KEY GOES HERE \
SPACES_REGION=ams3 \
SPACES_SPACE=SPACE NAME GOES HERE \
mdmw
```

### license

MIT. see [LICENSE](/LICENSE)
