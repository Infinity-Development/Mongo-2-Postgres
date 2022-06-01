package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/nickwells/pager.mod/pager"
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
	col                string
	tgtKey             string
	tgtVal             string
	filterHrs          int
)

type KVPair struct {
	Key   string
	Value any
}

func pageOutput(obj []KVPair) {
	pagerW := pager.W()

	s := pager.Start(&pagerW)

	// Send the objects here with a pager

	var text string

	for _, obj := range obj {
		text += fmt.Sprintln(obj.Key, "=>", obj.Value)
	}

	fmt.Fprintln(pagerW.StdW(), text)

	s.Done()
}

func uuidToString(id pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}

func analyzeBackup(mongoCtx context.Context, db *mongo.Database) {
	var (
		analyzeObjects [][]KVPair
		analyzeTs      []pgtype.Timestamptz
		analyzeId      []pgtype.UUID
	)

	ctx := context.Background()

	fmt.Println("DBTool: analyze")

	if col == "" || tgtKey == "" || tgtVal == "" {
		panic("No valid col/tgtKey/tgtVal found")
	}

	conn, err := pgx.Connect(ctx, backupDbName)

	if err != nil {
		panic(err)
	}

	interval := "interval '" + strconv.Itoa(filterHrs) + " hours'"

	query, err := conn.Query(ctx, "SELECT id, data, ts FROM backups WHERE col = $1 AND (NOW() - ts) < "+interval, col)

	if err != nil {
		panic(err)
	}

	defer query.Close()

	var i int
	var keysFound int
	var objectsFound int

	for query.Next() {
		var id pgtype.UUID
		var data pgtype.JSONB
		var ts pgtype.Timestamptz

		if err := query.Scan(&id, &data, &ts); err != nil {
			panic(err)
		}

		var encode []KVPair

		err := json.Unmarshal(data.Bytes, &encode)

		if err != nil {
			fmt.Println("JSON parse error:", err)
			continue
		}

		// Now we check if this is the target
		for _, kv := range encode {
			if kv.Key == tgtKey {
				keysFound++

				var isVal bool

				switch kv.Value.(type) {
				case string:
					if kv.Value.(string) == tgtVal {
						isVal = true
					}
				}

				if isVal {
					fmt.Println("Found object at ts", ts.Time)
					analyzeObjects = append(analyzeObjects, encode)
					analyzeTs = append(analyzeTs, ts)
					analyzeId = append(analyzeId, id)
					objectsFound++
					break
				}
			}
		}

		// Update i
		i++
	}

	reloadNeeded := false

	// Menu
	for {
		fmt.Println("\n\n\nANALYSIS OUTPUT\n===============")

		for i := range analyzeObjects {
			fmt.Println(strconv.Itoa(i+1)+". Found backup", uuidToString(analyzeId[i]), "at ts", analyzeTs[i].Time)
			fmt.Println()
		}

		fmt.Println("\nFound", i, "entities total of which", keysFound, "keys matching the target key were found")

		fmt.Println("A total of", objectsFound, "objects were found matching this criteria:")

		reader := bufio.NewReader(os.Stdin)

		// Print menu
		fmt.Println("\nE: Exit menu and return")
		fmt.Println("L: Look at a backup using pager")
		fmt.Println("R: Restore a backup")
		fmt.Println("D: Delete a backup")
		fmt.Println("DA: Deletes all above backups")

		fmt.Print("\n\nSelect an option: ")
		text, _ := reader.ReadString('\n')

		text = strings.ReplaceAll(text, "\n", "")

		switch text {
		case "E":
			os.Exit(0)
		case "L":
			fmt.Print("Which of the above ", objectsFound, " backups do you wish to view: ")
			backupNum, _ := reader.ReadString('\n')
			backupNum = strings.ReplaceAll(backupNum, "\n", "")

			// Parse backupNum to text
			backupId, err := strconv.Atoi(backupNum)
			if err != nil {
				fmt.Println("Error while parsing backup number:", backupId)
			} else {
				backupId = backupId - 1
				if len(analyzeObjects) <= backupId || backupId < 0 {
					fmt.Println("Error while parsing backup number: invalid number")
				} else {
					pageOutput(analyzeObjects[backupId])
				}
			}
		case "R":
			fmt.Print("Which of the above ", objectsFound, " backups do you wish to restore: ")
			backupNum, _ := reader.ReadString('\n')
			backupNum = strings.ReplaceAll(backupNum, "\n", "")

			// Parse backupNum to text
			backupId, err := strconv.Atoi(backupNum)
			if err != nil {
				fmt.Println("Error while parsing backup number:", backupId)
			} else {
				backupId = backupId - 1
				if len(analyzeObjects) <= backupId || backupId < 0 {
					fmt.Println("Error while parsing backup number: invalid number")
				} else {
					var backup map[string]any = make(map[string]any, len(analyzeObjects[backupId]))

					for _, val := range analyzeObjects[backupId] {
						if val.Key == "_id" {
							continue // Let mongo figure out _id
						}
						backup[val.Key] = val.Value
					}

					fmt.Println("Restoring backup of document to", col)

					// Delete old
					var filterCond = make(map[string]any)

					filterCond[tgtKey] = tgtVal

					delRes, err := db.Collection(col).DeleteMany(mongoCtx, filterCond)

					if err != nil {
						panic(err)
					}

					fmt.Println("Deleted", delRes.DeletedCount, " old entities")

					res, err := db.Collection(col).InsertOne(mongoCtx, backup)
					if err != nil {
						panic(err)
					}
					fmt.Println("Restored document ID:", res.InsertedID)
				}
			}
		case "DA":
			for _, v := range analyzeId {
				_, err := conn.Exec(ctx, "DELETE FROM backups WHERE id = $1", uuidToString(v))

				if err != nil {
					panic(err)
				}
			}

			fmt.Println("Successfully deleted all backup")
			reloadNeeded = true

		case "D":
			fmt.Print("Which of the above ", objectsFound, " backups do you wish to delete: ")
			backupNum, _ := reader.ReadString('\n')
			backupNum = strings.ReplaceAll(backupNum, "\n", "")

			// Parse backupNum to text
			backupId, err := strconv.Atoi(backupNum)
			if err != nil {
				fmt.Println("Error while parsing backup number:", backupId)
			} else {
				backupId = backupId - 1
				if len(analyzeObjects) <= backupId || backupId < 0 {
					fmt.Println("Error while parsing backup number: invalid number")
				} else {
					_, err := conn.Exec(ctx, "DELETE FROM backups WHERE id = $1", uuidToString(analyzeId[backupId]))

					if err != nil {
						panic(err)
					}

					fmt.Println("Successfully deleted backup")
					reloadNeeded = true
				}
			}
		default:
			fmt.Println("Invalid input", text, []byte(text))
		}

		fmt.Println("Retrying menu in 3 seconds.")
		time.Sleep(3 * time.Second)

		if reloadNeeded {
			break
		}
	}

	if reloadNeeded {
		analyzeBackup(mongoCtx, db)
	}
}

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
				if act != "watch" {
					continue
				}
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

	actExamples := []string{
		"Analyze a bot in a collection named bots to restore it back to DB: 'db-backup-tool --act analyze --col bots --tgtKey botID --tgtVal 678040985678118912 --filterHrs 12'",
		"Create a new backup: 'db-backup-tool --act backup'",
	}

	actExamplesStr := "\n\t" + strings.Join(actExamples, "\n\n\t") + "\n"

	flag.StringVar(&connString, "conn", "mongodb://127.0.0.1:27017/infinity", "[This is required] MongoDB connection string")
	flag.StringVar(&dbName, "dbname", "infinity", "[This is required] DB name to connect to.")
	flag.StringVar(&act, "act", "", "[This is required] Action to perform (backup/watch/analyze). If act is analyze, then col must be set to the collection/column to analyze, tgtKey to the key on the document to test for and tgtVal as the value on the document to validate against\n\nExamples:\n"+actExamplesStr)
	flag.StringVar(&backupDbName, "backup-db", "postgresql://127.0.0.1:5432/backups?user=root&password=iblpublic", "[This is required] Backup Postgres DB URL")
	flag.IntVar(&backupTimeInterval, "interval", 60, "[This is required if using act as watch] Interval for watcher to wait for (minutes)")
	flag.StringVar(&ignored, "ignore", "sessions", "[This is required] What collections to ignore, seperate using ,! Spaces are ignored. Use none to not ignore any collections")
	flag.StringVar(&col, "col", "", "[This is required if using act as analyze] Column/collection to analyze (analyze only)")
	flag.StringVar(&tgtKey, "tgtKey", "", "[This is required if using act as analyze] The key on the document to filter analysis on (as in, which key do you want to use to look up a backup) (analyze only)")
	flag.StringVar(&tgtVal, "tgtVal", "", "[This is required if using act as analyze] The value of the key specified in tgtKey to filter analysis on (as in, which value of the tgtKey should be used to lookup a backup. An analogy is mongo's db.findOne with one condition) (analyze only)")
	flag.IntVar(&filterHrs, "filterHrs", 4, "[Optional] How many hours to look back during analyze (analyze only)")

	flag.Parse()

	if ignored != "none" {
		ignoredCols = strings.Split(strings.ReplaceAll(ignored, " ", ""), ",")
	}

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
	} else if act == "analyze" {
		// This is the anaylzer
		analyzeBackup(ctx, db)
	} else {
		panic("Unsupported operation")
	}
}
