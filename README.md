# Mongo-2-Postgres
Simple cli used to backup and migrate data from Mongoose to Postgres on a Rotating/Rolling basis.

--- 

## Setup
• Check back soon!

---

## Usage

| Command      | Description                                                   | Example                                        |
| ------------ | ------------------------------------------------------------- | ---------------------------------------------- |
| `help `      | List available commands and their usage                       | `./db-backup-tool --help`                      |
| `act`        | Specify an action to perform (backup/watch)                   | `./db-backup-tool --act watch`                 |
| `backup-db`  | Specify the Postgres Connection String.                       | `./db-backup-tool --backup-db POSTGRES_URL`    |
| `conn`       | Specify the Mongoose Connection String.                       | `./db-backup-tool --conn MONGO_URL`            |
| `dbname`     | Specify the Mongoose Database Name.                           | `./db-backup-tool --dbname MONGO_NAME`         |
| `interval`   | Interval for watcher to wait for (default 60 mins).           | `./db-backup-tool --interval SOME_INT`         |

---

## Example Watch Output

```
DBTool: init
DBTool: Connecting to mongodb://127.0.0.1:27017/infinity
Connected to mongoDB?
Collections in DB:  [packages staff_apps dev_apps users reviews transcripts sessions tickets oauths suggests bots votes]
DBTool: Connected to mongo successfully
Backing up packages
Backing up staff_apps
Backing up dev_apps
Backing up users
Backing up reviews
Backing up transcripts
Ignoring sessions as it is in ignoredCols
Backing up tickets
Backing up oauths
Backing up suggests
Backing up bots
Backing up votes
Waiting for next backup rotation
```

## Contributors
<a href="https://github.com/InfinityBotList/Mongo-2-Postgres/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=InfinityBotList/Mongo-2-Postgres" />
</a>
