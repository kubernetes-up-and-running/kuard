A demo application for "Kubernetes Up and Running".


### Building

Have Go installed.

```
go get -u github.com/jteeuwen/go-bindata/...
make push
```

### Makefiles

Go building makefiles taken from
https://github.com/thockin/go-build-template with an Apache 2.0 license.
Handling multiple targets taken from https://github.com/bowei/go-build-template.

### TODO
* [ ] Make file system browser better.  Show size, permissions, etc.  Might be able to do this by faking out an `index.html` as part of the http.FileSystem stuff.
* [ ] Switch to Angular with an API to the server.  This mixed template stuff sucks.
