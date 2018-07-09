This is a simple CRUD web app. It is very only used to test / practice infra deployments

this project uses `dep` install from https://github.com/golang/dep

To build `go build` or to cross compile got all support os's `gox` or for linux if on the Mac, `gox -osarch="linux/amd64"`

Config is taken from env vars, all logging is to stdout.


Run like this:

`EXAMPLE_APP_DB_HOST=localhost EXAMPLE_APP_DB_PORT=5432 EXAMPLE_APP_DB_NAME=mydb EXAMPLE_APP_DB_USER=postgres EXAMPLE_APP_DB_PASSWD=example ./example_crud`

If anyone does use this and finds issues, happy to take pull requests.