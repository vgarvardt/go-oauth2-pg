# PostgreSQL Storage for [OAuth 2.0](https://github.com/go-oauth2/oauth2)

[![Build][Build-Status-Image]][Build-Status-Url] [![Codecov][codecov-image]][codecov-url] [![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

## Install

```bash
$ go get -u -v github.com/vgarvardt/go-oauth2-pg
```

## PostgreSQL drivers

The store accepts an adapter interface that interacts with the DB. Adapter and implementations are extracted to separate package [`github.com/vgarvardt/go-pg-adapter`](https://github.com/vgarvardt/go-pg-adapter) for easier maintenance.

## Usage example

```go
package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	pg "github.com/vgarvardt/go-oauth2-pg"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
	"gopkg.in/oauth2.v3/manage"
)

func main() {
	pgxConn, _ := pgx.Connect(context.TODO(), os.Getenv("DB_URI"))

	manager := manage.NewDefaultManager()

	// use PostgreSQL token store with pgx.Connection adapter
	adapter := pgx4adapter.NewConn(pgxConn)
	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
	defer tokenStore.Close()
	
	clientStore, _ := pg.NewClientStore(adapter)

	manager.MapTokenStorage(tokenStore)
	manager.MapClientStorage(clientStore)
	// ...
}
```

## How to run tests

You will need running PostgreSQL instance. E.g. the one running in docker and exposing a port to a host system

```bash
docker run --rm -p 5432:5432 -it -e POSTGRES_PASSWORD=oauth2 -e POSTGRES_USER=oauth2 -e POSTGRES_DB=oauth2 postgres:10
```

Now you can run tests using the running PostgreSQL instance using `PG_URI` environment variable

```bash
PG_URI=postgres://oauth2:oauth2@localhost:5432/oauth2?sslmode=disable go test -cover ./...
```

## MIT License

```
Copyright (c) 2019 Vladimir Garvardt
```

[Build-Status-Url]: https://travis-ci.org/vgarvardt/go-oauth2-pg
[Build-Status-Image]: https://travis-ci.org/vgarvardt/go-oauth2-pg.svg?branch=master
[codecov-url]: https://codecov.io/gh/vgarvardt/go-oauth2-pg
[codecov-image]: https://codecov.io/gh/vgarvardt/go-oauth2-pg/branch/master/graph/badge.svg
[reportcard-url]: https://goreportcard.com/report/github.com/vgarvardt/go-oauth2-pg
[reportcard-image]: https://goreportcard.com/badge/github.com/vgarvardt/go-oauth2-pg
[godoc-url]: https://godoc.org/github.com/vgarvardt/go-oauth2-pg
[godoc-image]: https://godoc.org/github.com/vgarvardt/go-oauth2-pg?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg
