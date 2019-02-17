# PostgreSQL Storage for [OAuth 2.0](https://github.com/go-oauth2/oauth2)

## Install

```bash
$ go get -u -v github.com/vgarvardt/go-oauth2-pg
```

## Usage

```go
package main

import (
	"gopkg.in/oauth2.v3/manage"
	pg "github.com/vgarvardt/go-oauth2-pg"
)

func main() {
	manager := manage.NewDefaultManager()

	// use mysql token store
	store := mysql.NewDefaultStore(
		mysql.NewConfig("root:123456@tcp(127.0.0.1:3306)/myapp_test?charset=utf8"),
	)

	defer store.Close()

	manager.MapTokenStorage(store)
	// ...
}

```

## MIT License

```
Copyright (c) 2018 Vladimir Garvardt
```
