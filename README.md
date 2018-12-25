# QRCodeApi
QRCodeApi is A REST API Server that generates QR code.written in Golang.
Response QR Code images.

please keep simple, small, smart. thanks.

# Status
Development only.

# Requirement
go1.11 or later(including `$GOPATH` setup).

# Installation

clone the source code by running the following command.
```
$ mkdir -p $GOPATH/src/github.com/stakada7
$ cd $GOPATH/src/github.com/stakada7
$ git clone git@github.com:stakada7/QrCodeApi.git
```

To fetch dependencies and build, run the following make tasks,
```
make
```

# Usage

To run `qrcodeapi`
```
$ bin/{YOUR_OS}/{YOUR_ARCH}/{VERSION}/qrcodeapi
```

# Specification

- `GET /`

Sample API. Development only. Response Sample QRCode.


- `GET /ping`

Health Check endpoint.

- `POST /`

Create QRCode API. You Should add HEADER and BODY.

```
curl -X POST http://localhost:3333/ -H 'authorization: Bearer XXXXXXXXXXXXXXXXXXXXXXXXX.XXXXXXXXXXXXXXXXXXXX' -d '{"Url":"https://www.youtube.com/watch?v=EC0BvUaD_Rk"}'
```

- `GET /list`

```
curl http://localhost:3333/list -H 'authorization: Bearer XXXXXXXXXXXXXXXXXXXXXXXXX.XXXXXXXXXXXXXXXXXXXX'
```

# benchmark

you should use `apib`. `apib` is benchmark tool.

```
apib -c 10 -d 10 -f body -x POST -t 'application/json' -H 'authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOjEyM30.Dd5OHdR0q32rc5SQEroras2j8m4DUmMYuNpjrsUTW6E' http://localhost:3333/
```
