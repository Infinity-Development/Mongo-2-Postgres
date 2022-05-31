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

## Contributors
<a href="https://github.com/InfinityBotList/Mongo-2-Postgres/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=InfinityBotList/Mongo-2-Postgres" />
</a>
