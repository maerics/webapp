# webapp

A Go backend web application skeleton with opinionated logging and basic database setup. Bring your own frontend framework.


```
├── cmd            - command line actions.
├── db             - database utilities.
│   └── migrations - SQL files for db migrations.
├── models         - go structs for db (and general) models.
├── public         - frontend asset output folder (git ignored).
└── web            - web server, config, and routing.
```

## Setup

1. Export this repository:
   ```sh
   git clone https://github.com/maerics/webapp
   cd webapp
   ```
1. Globally replace the token `webapp` with your app name, e.g. "myapp":
   ```sh
   sed -i '' s/webapp/myapp/g $(git grep -lw webapp)
   ```
1. Run `make` to ensure tests pass and remove the git directory:
   ```
   make && rm -rf .git
   ```
1. Remove this "Setup" section from the README.

## Overview

1. Forward-only, idempotent database migrations go in `db/migrations/`
   ```sh
   go run . db migrate
   ```
1. Custom backend routes go in `web/routes.go`.
1. Run the webserver on [http://localhost:8080](http://localhost:8080)
   ```sh
   go run . web
   ```
1. Frontend assets go in `public/` and are hot-reloaded by default.
1. Running with `GIN_MODE=release` serves embedded assets in `public/` at build time for a single portable executable.
1. Custom command line functions go in `cmd`
   ```sh
   go run .     # Print help message for all commands
   go run . db  # Print help message for "db" commands
   ```
