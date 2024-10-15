# super-invoicer

See https://github.com/upsidr/coding-test/blob/main/web-api-language-agnostic/README.ja.md

## Help

```console
$ go run main.go -h
App for creating and getting invoices.

Usage:
   [flags]

Flags:
  -h, --help   help for this command
```

## Environment Variables

| Name                  | Description                                                                       |
| --------------------- | --------------------------------------------------------------------------------- |
| `MYSQL_USERNAME`      | Username for mysql cluster                                                        |
| `MYSQL_PASSWORD`      | Password for mysql cluster                                                        |
| `MYSQL_ROOT_PASSWORD` | Root password for mysql cluster (NOTE: necessary only when run in docker compose) |

## How to run in docker compose environment

```console
$ MYSQL_USERNAME=root
$ MYSQL_PASSWORD=mysql
$ MYSQL_ROOT_PASSWORD=mysql
$ make up
```
