# Local OAuth2

This is a utility that allows you to grab an id_token from auth0 from the
command line. It requires that you have created an application in the auth0
dashboard, and you have set it up to allow a callback url of http://localhost:10000.

Once that is all working you can do this.

```
export ID_TOKEN=$(_work/local_oauth2_darwin_amd64 -clientid=sdfsdfsdf -host my-server.auth0.com)
```

## Building

To start

```
make init
```

To test

```
make test
```

to build

```
make build
```
