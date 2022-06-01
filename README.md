<h2 align='center'>
  <img src="https://media.discordapp.net/attachments/653733403841134600/981292240137769001/IMG_5344.png" height='100px' width='100px' />
  <br> 
  Mongo to PostgreSQL
</h2>
<p align="center">
   Simple cli used to backup and migrate data from MongoDB (Mongoose etc.) to Postgres on a Rotating/Rolling basis.
</p>

[![Version](https://img.shields.io/badge/Version-1.0.1%20-green.svg?style=flat)](https://github.com/InfinityBotList/Mongo-2-Postgres) 
[![Made with](https://img.shields.io/badge/Language-GO%20-blue.svg?style=flat)](https://github.com/InfinityBotList/Mongo-2-Postgres) 
[![License Type](https://img.shields.io/badge/License-MIT-yellowgreen.svg)](https://github.com/InfinityBotList/Mongo-2-Postgres)

--- 

## Setup

This tool requires golang 1.18 (older versions *may* work with some patches but this is *not* supported)

For ease of use, this tool comes with a ``Makefile`` (which just invokes ``go build``). The ``go install
`` command can be used to install the binary to ``$GOPATH/bin`` like all go programs. You can also run ``make install`` which will copy the built binary to ``/usr/bin`` (Linux only) if you want this for any specific reason.

As a backup source, postgres is currently required, **feel free to make a Pull Request if you wish to add support for other databases**

As a source, this tool was created specifically for mongoDB, however **feel free to make a Pull Request if you wish to add support for another database**

A mongoDB connection string can be provided via the ``--conn`` command line argument. Similarly, the postgreSQL connection string can be provided via the ``backup-db`` command line argument.

This tool should be self-explanatory to use. Feel free to make a Issue if you have any problems using it!

---

## Features
- Built-In Watcher to allow for interval based back ups.
- Ability to ignore specified collections/models as a whole. 
- Easy to use, setup and configure. Extremely performant and reliable.

---

## Usage

You can run ``./db-backup-tool --help`` for help on the options provided

This tool supports **single backups** and **long-running automatic backups**. The long-running backups are made through a "watcher" action/mode and once run in this way can then be daemonized using systemd/tmux etc.

It was originally made to help allievate issues surrounding Infinity Bot Lists database which even ZFS zvol + XFS failed to resolve which is likely due to unmaintainable code by the prior owner (which is being rewritten at this time). It is provided as a open source software in the hopes that others will benefit from it.

### Debug Stats

This tools provides debug stats via the signal ``USR1`` (user-defined signal 1).

Example: ``pkill -USR1 db-backup-tool``

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
