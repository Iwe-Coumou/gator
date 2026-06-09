# gator

A CLI RSS feed aggregator written in Go. Subscribe to feeds, aggregate posts in the background, and browse your latest content from the terminal.

## Prerequisites

- Go 1.25+
- PostgreSQL

## Installation

```bash
go install github.com/Iwe-Coumou/gator@latest
```

Or build from source:

```bash
git clone https://github.com/Iwe-Coumou/gator
cd gator
go build -o gator .
```

## Configuration

Create `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Apply the database schema using [goose](https://github.com/pressly/goose) or `psql`:

```bash
psql $DATABASE_URL -f sql/schema/001_users.sql
psql $DATABASE_URL -f sql/schema/002_feeds.sql
psql $DATABASE_URL -f sql/schema/003_feed_follows.sql
psql $DATABASE_URL -f sql/schema/004_last_fetched.sql
psql $DATABASE_URL -f sql/schema/005_posts.sql
```

## Usage

### User management

```bash
gator register <username>   # Create a new user and set them as current
gator login <username>      # Switch the current user
gator users                 # List all users (* marks the current user)
gator currentUser           # Print the current user
gator reset                 # Delete all users (destructive)
```

### Feed management

```bash
gator addfeed <name> <url>  # Add a feed and follow it (requires login)
gator feeds                 # List all feeds
gator follow <url>          # Follow an existing feed (requires login)
gator following             # List feeds followed by the current user
gator unfollow <url>        # Unfollow a feed (requires login)
```

### Aggregation and browsing

```bash
gator agg <interval>        # Poll all feeds on an interval (e.g. 30s, 5m, 1h)
gator browse [limit]        # Show the latest posts for the current user (default: 2)
```

`agg` runs continuously until interrupted — run it in a separate terminal or as a background process while using the other commands.

## Project structure

```
gator/
├── main.go                  # Entry point, wires state and registers commands
├── commands.go              # Command dispatch
├── handlers.go              # Command handlers and feed scraping logic
├── rss.go                   # RSS fetching and XML parsing
├── internal/
│   ├── config/              # Config file read/write (~/.gatorconfig.json)
│   └── database/            # sqlc-generated database layer
└── sql/
    ├── queries/             # SQL queries used by sqlc
    └── schema/              # Migration files
```
