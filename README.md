# Gator - RSS Feed Aggregator CLI

Gator is a command-line RSS feed aggregator project built in Golang. It allows you to add, follow and browse RSS feeds from the terminal. Feeds are stored in PostgreSQL, and the aggregator runs in the background, continuously fetching new posts.

-----
# Pre-requisites:

To run Gator, you will need to install:
- [Go](https://go.dev/doc/install) (1.26+)
- [PostgreSQL](https://www.postgresql.org/download/)

Ensure PostgreSQL is running and you have a database created:

```
sql
CREATE DATABASE gator;
```

-----
# Install guide:

```bash
go install github.com/devjoshbrown/gator@latest
```
This will install the `gator` binary to your `$GOPATH/bin` directory. Make sure that directory is in your `PATH`.

-----
# Config:
Create a config file at `~/.gatorconfig.json` with the following content:

```
json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator",
  "current_user_name": ""
}
```

Update the `db_url` to match your PostgreSQL connection string if it differs from the default.

-----
# Database Migrations

Install [Goose](https://github.com/pressly/goose) for running migrations:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Run the migrations:

```bash
goose -dir sql/schema postgres "your_db_url_here" up
```



## Usage ## 

-----
# User Management

```bash
gator register <name>    # Register a new user
gator login <name>       # Login as an existing user
gator users              # List all users
```

-----
# Feed Management

```bash
gator addfeed <name> <url>   # Add a new feed (automatically follows it)
gator feeds                  # List all feeds
```

-----
# Following

```bash
gator follow <url>       # Follow an existing feed
gator unfollow <url>     # Unfollow a feed
gator following          # List feeds you're following
```

-----
# Aggregation

Start fetching posts continuously in a separate terminal:

```bash
gator agg 30s            # Fetch every 30 seconds
```

Use `Ctrl+C` to stop.

-----
# Browsing Posts

```bash
gator browse             # Browse posts (default: 2 posts)
gator browse 10          # Browse with a custom limit
```

-----
# Reset

```bash
gator reset              # Delete all data (use with caution)
```

-----
# Example RSS Feeds

- Boot.dev Blog: `https://www.boot.dev/blog/index.xml`
- TechCrunch: `https://techcrunch.com/feed/`
- Hacker News: `https://news.ycombinator.com/rss`
- Wags Lane: `https://www.wagslane.dev/index.xml`

Thanks for checking out my Blog_Aggrigator!
Developed by Josh Brown [DevJoshBrown]
