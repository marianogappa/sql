# sql
MySQL pipe

## What does it do?

- `sql` allows you to pipe STDIN (hopefully containing SQL) to one or more pre-configured MySQL databases
- output comes out in `\t`-separated format, allowing further piping (e.g. works really well with [chart](https://github.com/MarianoGappa/chart))
- when more than one database is queried, the requests are made in parallel
- `sql` can either run `mysql` locally, run `mysql` locally but connecting to a remote host (by configuring a `dbServer`), or `ssh` to a remote host and from there run `mysql` to either a local or remote host (by configuring an `appServer` and a `dbServer`)

## Installation

Get the latest binary on the [Releases](https://github.com/MarianoGappa/sql/releases) section, or via `go get`:
```
go get -u github.com/MarianoGappa/sql
```

## Configuration

Create a `.databases.json` dotfile in your home folder. [This](.databases.json.example) is an example file.

## Example usages

```
cat query.sql | sql test_db

sed 's/2015/2016/g' query_for_2015.sql | sql db1 db2 db3

echo "SELECT * FROM users WHERE name = 'John'" | sql all
```

## Notes

- when more than one database is queried, the resulting rows are prefixed with the database identifier
- the `all` special keyword means "sql to all configured databases"
- `sql` assumes that you have correctly configured SSH keys on all servers you `ssh` to
- please note that `~/.databases.json` will contain your database credentials in plain text; if this is a problem for you, don't use `sql`!
- `sql` is meant for automation of one-time lightweight ad-hoc `SELECT`s on many databases at once; it's not recommended for mission critical bash scripts that do destructive operations on production servers!

## Dependencies

- mysql
- ssh (only if you configure an "appServer")

## Contribute

Yes, please.
