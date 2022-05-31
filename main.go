package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	connString         string
	dbName             string
	backupDbName       string
	act                string
	backupTimeInterval int
	ignoredCols        []string
	lastRotation       int
)

func backupDb(ctx context.Context, db *mongo.Database, colNames []string) {
	bkCtx := context.Background()

	conn, err := pgx.Connect(bkCtx, backupDbName)

	if err != nil {
		panic(err)
	}

	for _, column := range colNames {
		var flag bool
		for _, col := range ignoredCols {
			if col == column {
				fmt.Println("Ignoring", col, "as it is in ignoredCols")
				flag = true
			}
		}

		if flag {
			continue
		}

		fmt.Println("Backing up", column)
		col := db.Collection(column)
		cur, err := col.Find(ctx, bson.D{})
		if err != nil {
			panic(err)
		}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			raw := cur.Current
			var dataIface interface{}
			err := bson.Unmarshal([]byte(raw), &dataIface)
			if err != nil {
				panic(err)
			}
			_, err = conn.Exec(bkCtx, "INSERT INTO backups (col, data) VALUES ($1, $2)", column, dataIface)
			if err != nil {
				panic(err)
			}
		}
	}
}

func handleMaintSignals() {
	ch := make(chan os.Signal, 1)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGUSR1:
				tsl := time.Duration(int(time.Now().Unix())-lastRotation) * time.Second
				nextRotation := time.Duration(backupTimeInterval)*time.Minute - tsl
				fmt.Println("[DEBUG] lastRotation:", lastRotation, "| Time since last rotation:", tsl, "| Estimated time till next rotation:", nextRotation)
			}
		}
	}()
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2)
}

func main() {
	handleMaintSignals()

	fmt.Println("DBTool: init")
	ctx := context.Background()

	var ignored string

	flag.StringVar(&connString, "conn", "mongodb://127.0.0.1:27017/infinity", "MongoDB connection string")
	flag.StringVar(&dbName, "dbname", "infinity", "DB name to connect to")
	flag.StringVar(&act, "act", "", "Action to perform (backup/watch)")
	flag.StringVar(&backupDbName, "backup-db", "postgresql://127.0.0.1:5432/backups?user=root&password=iblpublic", "Backup Postgres DB URL")
	flag.IntVar(&backupTimeInterval, "interval", 60, "Interval for watcher to wait for (minutes)")
	flag.StringVar(&ignored, "ignore", "sessions", "What collections to ignore, seperate using ,! Spaces are ignored")

	flag.Parse()

	ignoredCols = strings.Split(strings.ReplaceAll(ignored, " ", ""), ",")

	progName := os.Args[0]

	if act == "" {
		fmt.Println("No action found. Try running:", progName, "--help")
		os.Exit(-1)
	}

	fmt.Println("DBTool: Connecting to", connString)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))

	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to mongoDB?")

	db := client.Database(dbName)

	colNames, err := db.ListCollectionNames(ctx, bson.D{})

	if err != nil {
		panic(err)
	}

	fmt.Println("Collections in DB: ", colNames)

	fmt.Println("DBTool: Connected to mongo successfully")

	if act == "backup" {
		backupDb(ctx, db, colNames)
	} else if act == "watch" {
		func() {
			d := time.Duration(backupTimeInterval) * time.Minute
			backupDb(ctx, db, colNames)
			fmt.Println("Waiting for next backup rotation")
			lastRotation = int(time.Now().Unix())
			for x := range time.Tick(d) {
				fmt.Println("Autobackup started at", x)

				colNames, err := db.ListCollectionNames(ctx, bson.D{})

				if err != nil {
					fmt.Println("Skipping backup as ListCollectionNames returned error:", err)
					continue
				}

				fmt.Println("Current collections in DB: ", colNames)

				backupDb(ctx, db, colNames)
				fmt.Println("Waiting for next backup rotation")
				lastRotation = int(time.Now().Unix())
			}
		}()
	} else {
		panic("Unsupported operation")
	}
}
