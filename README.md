A demo application for "Kubernetes Up and Running".


### Building

Have Go installed.

```
go get -u github.com/jteeuwen/go-bindata/...
make push
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

### Makefiles

Go building makefiles taken from
https://github.com/thockin/go-build-template with an Apache 2.0 license.
Handling multiple targets taken from https://github.com/bowei/go-build-template.

### TODO
* [ ] Make file system browser better.  Show size, permissions, etc.  Might be able to do this by faking out an `index.html` as part of the http.FileSystem stuff.
* [ ] Switch to Angular with an API to the server?  This mixed template stuff sucks.
* [ ] Find a better way to pick unique colors to show version.  Perhaps just have a table based on the fakeversion.
