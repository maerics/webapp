# webapp

A go web application skeleton with opinionated logging and database setup.

```
├── cmd            - command line actions.
├── db             - database utilities.
│   └── migrations - SQL files for db migrations.
├── models         - go structs for db (and general) models.
├── public         - frontend asset output folder (git ignored).
└── web            - web server, config, and routing.
```

## Setup

1. Clone this repository.
1. Globally replace the token `webapp` with your app name.
1. Database migrations go in `db/migrations/` then `go run . db migrate`.
1. Frontend assets go in `public/` then `go run . web` (hot reload).
1. Setting `GIN_MODE=release` embeds assets from `public/` (no hot reload).
1. Custom backend routes go in `web/routes.go`.
1. Browse to [http://localhost:8080](http://localhost:8080).
