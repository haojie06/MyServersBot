package main

import "log"

func checkError(err error, info ...string) {
	if err != nil && len(info) > 0 {
		log.Panic(info[0] + err.Error())
	} else if err != nil {
		log.Panic(err.Error())
	}
}
