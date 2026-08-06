# list-service

Watch lists for weeb.vip — what each user is watching, has finished, or has put
aside for later.

A Go GraphQL service built with [gqlgen](https://gqlgen.com), federated into the
gateway schema. MySQL through GORM, with changes published to Apache Pulsar so
other services can react to a list changing without this one knowing about them.

## Running it

Requires Go and MySQL.

```sh
make migrate                  # bring the database up to date
go run cmd/main.go server     # the GraphQL server
```

`config/config.dev.json` is the local config, overridden by environment
variables — see `config/config.go`.

## Schema and generated code

```sh
make gql        # regenerate resolvers from the schema
make mocks      # regenerate test mocks
make generate   # both
```

Both outputs are committed.

## Migrations

```sh
make create-migration name=add_something
make migrate
```

The migration history is worth reading in order: the list states were renamed
once (`plantowatch` became `watchlist`), and the up/down pair for it is in
`db/migrations`.

## Metrics

Through [go-metrics-lib](https://github.com/weeb-vip/go-metrics-lib), reporting
to Prometheus and Datadog.
