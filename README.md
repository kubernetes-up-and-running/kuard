A demo application for "Kubernetes Up and Running".


### Building

Have Docker installed.

```
make push REGISTRY=<my-gcr-registry>
```

### Versions

Images built will automatically have the git verison (based on tag) applied.  In addition, there is an idea of a "fake version".  This is used so that we can use the same basic server to demonstrate upgrade scenarios.

Right now we create 3 fake versions: `1`, `2`, and `3`.  This translates into the following container images:

```
gcr.io/kuar-demo/kuard-amd64:v0.2-1
gcr.io/kuar-demo/kuard-amd64:1
gcr.io/kuar-demo/kuard-amd64:v0.2-2
gcr.io/kuar-demo/kuard-amd64:2
gcr.io/kuar-demo/kuard-amd64:v0.2-3
gcr.io/kuar-demo/kuard-amd64:3
```

For documentation where you want to demonstrate using versions but use the latest version of this server, you can simply reference `gcr.io/kuar-demo/kuard-amd64:1`.  You can then demonstrate an upgrade with `gcr.io/kuar-demo/kuard-amd64:2`.

(Another way to think about it is that `:1` is essentially `:latest-1`)

### Development

If you just want to do Go server development, you can build the client as part of a build `make`.  It'll drop the result in to `sitedata/built/`.

If you want to do both Go server and React.js client dev, you need to do the following:
1. Have Node installed
2. In one terminal
  * `cd client`
  * `npm install`
  * `npm start`
  * This will start a debug node server on `localhost:8081`.  It'll proxy all unhandled requests to `localhost:8080`
3. In another terminal
  * `go generate ./...`
  * `go run cmd/kuard/*.go --debug`
4. Open your browser to http://localhost:8081.

This should support live reload of any changes to the client.  The Go server will need to be exited and restarted to see changes.

### Makefiles

Go building makefiles taken from
https://github.com/thockin/go-build-template with an Apache 2.0 license.
Handling multiple targets taken from https://github.com/bowei/go-build-template.

These have been heavily modified.
* Support explicit docker volume for caching vs. using host mounts (as they are really slow on macOS)
* Building/caching node
* Fake versions so we can play with upgrades of this server

### TODO
* [ ] Make file system browser better.  Show size, permissions, etc.  Might be able to do this by faking out an `index.html` as part of the http.FileSystem stuff.
* [x] Switch to React with an API to the server?  This mixed template stuff sucks.
