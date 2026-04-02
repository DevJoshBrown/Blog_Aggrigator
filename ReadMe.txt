The 3 Systems and What They Do

| System
| Where
| What it does


| **Config**
| `internal/config/`
| Reads/writes `~/.gatorconfig.json`. Just stores your current username and DB URL

| **Database (SQLC)**
| `sql/queries/` → `internal/database/`
| You write SQL, SQLC generates Go functions. You never edit `internal/database/`

| **Commands**
| `main.go`
| Each command is a function that receives `state` and does something with the config and/or database

## How Adding a Feature Always Works

Every new feature follows the exact same 3 steps:

Step 1 — Write the SQL in `sql/queries/users.sql`
Step 2 — Run `sqlc generate` to create the Go function
Step 3 — Write a handler in `main.go` that calls that Go function

That's it. Every. Single. Time.
