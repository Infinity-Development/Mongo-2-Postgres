package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	connString   string
	dbName       string
	backupDbName string
	act          string
)

func backupDb(ctx context.Context, db *mongo.Database, colNames []string) {
	bkCtx := context.Background()

	conn, err := pgx.Connect(bkCtx, backupDbName)

	if err != nil {
		panic(err)
	}

	for _, column := range colNames {
		if column == "sessions" {
			fmt.Println("Ignoring sessions as it is way too big")
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

func main() {
	fmt.Println("DBTool: init")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	flag.StringVar(&connString, "conn", "mongodb://127.0.0.1:27017/infinity", "MongoDB connection string")
	flag.StringVar(&dbName, "dbname", "infinity", "DB name to connect to")
	flag.StringVar(&act, "act", "", "Action to perform")
	flag.StringVar(&backupDbName, "backup-db", "postgresql://127.0.0.1:5432/backups?user=root&password=iblpublic", "Backup Postgres DB URL")

	flag.Parse()

	if act == "" {
		fmt.Println("No action found")
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

	fmt.Println("Collections in DB: ", colNames)

	fmt.Println("DBTool: Connected to mongo successfully")

	if act == "backup" {
		backupDb(ctx, db, colNames)
	} else {
		panic("Unsupported operation")
	}
}