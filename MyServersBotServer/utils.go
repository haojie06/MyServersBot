package main

import "log"

func checkError(err error, info ...string) {
	if err != nil {
		log.Panic(info[0] + err.Error())
	}
}
