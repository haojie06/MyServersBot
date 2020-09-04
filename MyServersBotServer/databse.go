package main

import "github.com/syndtr/goleveldb/leveldb"

func startDB() *leveldb.DB {
	db, err := leveldb.OpenFile("db/botdb", nil)
	checkError(err)
	return db
}

func closeDB(db *leveldb.DB) {
	err := db.Close()
	checkError(err)
}
