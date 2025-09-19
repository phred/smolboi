[![Build and push to ghcr.io](https://github.com/phred/smolboi/actions/workflows/ghcr-publish.yml/badge.svg)](https://github.com/phred/smolboi/actions/workflows/ghcr-publish.yml)

# 📼 smolboi 📼

A mixtape 📼 of the greatest hits from [goStatic](https://github.com/PierreZ/goStatic) with a little special sauce all my my own.

## The goals
- Wafer thin OCI image
- Aerogel weight Nix build
- Handle the boring-ass job of serving [my website](https://fff.red) to my 2 visitors a week (a guess, thanks for reading if you do!). At present I don't track metrics on this and I hope I have >0 visitors.

The original project's goal was "to create to smallest docker container for my web static files." I'm tired of Docker and want something smaller, and simpler, if I can get it. What is a Docker image but an executable tarfile?

So far I have added virtualhost support and have been running my site on it for a couple years now.

I run this on `fly.io` with a container image I push to GitHub's Container repository. A ~5mb web server layer combined with a ~17mb static file layer makes for very quick builds and deploys. There's not much to go wrong.

### Wait, can I use this code?

Go for it, take my changes and make your own remix. What you do with that code is up to you. I'm not going to add any features that aren't related to making my webspace cooler.


## Features
 * A fully static web server embedded in a `SCRATCH` image
 * No framework
 * Web server built for Docker
 * Light container
 * 🆕 Virtual Hosting
 * 🆕 custom 404 pages--create a page called `404.html` at the site root, it gets served for [missing pages](https://fff.red/thispageisnothere).

Deleted or remixed features:
 * ~~Custom response headers per path and filetype~~
 * ~~Basic authentication~~
 * ~~Optional~~ Healthcheck--`/health` is always enabled.
 * 🆕 Request logs always enabled
 * 🆕 nix flake -- try `nix run github:phred/smolboi -- -path ~/to/your/static/public_html`


## Why?
Because Caddy2 is too complex, and I don't want to configure yet another instance of Apache or Nginx. I like Deno, but I like Go's production-grade standard library HTTP server better. Rust is cool but makes fat binaries too. Go also makes fat binaries, but this project has a minimal [set of dependencies](./go.mod)

The world is full of terrible, bloated software, fat containers, `node_modules` by the truckload. Big binaries have hidden externalities in transport, storage, compute time, energy costs. My needs are simple and they fit in a single binary, [as things should](https://fossil-scm.org/).

Thank you, PierreZ, for showing me the way. Software complexity will eat us all, best to only use what you need for the job.

## How to use
```
docker run -d -p 80:8043 -v /path/to/website:/srv/http --name smolboi ghcr.io/phred/smolboi:latest
```

## Usage 

```
Usage of ./smolboi:
  -log-level string
      default: info - What level of logging to run, info logs all requests (error, warn, info, debug) (default "info")
  -path string
      The path for the static files (default "/srv/http")
  -port int
      The listening port (default 8043)
  -vhost string
      The prefix for locating lightweight virtual hosted subdomains, or vhosts. E.g. 'labs' will serve the files at /srv/http/labs/tango when someone visits http://tango.your.tld (default "labs")
```


## Build

### I think this works, give it a whirl
```bash
docker buildx create --use --name=cross
docker buildx build --platform=linux/amd64,linux/arm64,linux/arm/v5,linux/arm/v6,linux/arm/v7,darwin/amd64,darwin/arm64,windows/amd64 .
```

I got tired of running Docker on my Mac so I let GitHub Workflows build the phred/smolboi containers. My [current container workflow](https://github.com/phred/smolboi/blob/trunk/.github/workflows/ghcr-publish.yml#L49C22-L49C58) builds only `linux/amd64,linux/arm64,darwin/arm64`.

### nix
This includes a `flake.nix` to get started. `nix run`, `nix develop`, `nix shell` all work as expected.
