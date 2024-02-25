# webapp

A Go backend web application skeleton with opinionated logging and PostgreSQL database connectivity. Bring your own frontend framework.


```
├── cmd            - command line actions.
├── db             - database utilities.
│   └── migrations - SQL files for db migrations.
├── models         - types for db and general models.
└── web            - web server, config, and routing.
    └── public     - static frontend assets.
```

## Setup

1. Export this repository:
   ```sh
   git clone https://github.com/maerics/webapp
   cd webapp
   ```
1. Globally replace the string "webapp" with your app name, e.g. "myapp":
   ```sh
   sed -i '' s/webapp/myapp/g $(git grep -lw webapp)
   ```
1. Run `make` to ensure tests pass (and reset git):
   ```
   export TEST_DATABASE_URL=postgres://webapp:p@localhost:5432/webapp_test
   make \
     && rm -rf .git \
     && git init    \
     && git add .   \
     && git commit -m 'Initial commit'
   ```
1. Remove this "Setup" section from the README.

## Overview

1. Export database connection variables
   ```sh
   export DATABASE_URL=postgres://webapp:p@localhost:5432/webapp_dev
   export TEST_DATABASE_URL=postgres://webapp:p@localhost:5432/webapp_test
   ```
1. Forward-only, idempotent database migrations go in `db/migrations/`
   ```sh
   go run . db migrate
   ```
1. Database seeding via
   ```sh
   go run . db seed
   ```
1. Custom backend routes go in `web/routes.go`.
1. Run the webserver on [http://localhost:8080](http://localhost:8080)
   ```sh
   go run . web
   ```
1. Frontend assets go in `web/public/` and are hot-reloaded by default.
1. Running with `MODE=release` serves embedded assets in `web/public/` at build time for a single portable executable.
1. Custom command line functions go in `cmd`
   ```sh
   go run .     # Print help message for all commands
   go run . db  # Print help message for "db" commands
   go run . web # Run the web server
   ```
