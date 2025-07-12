package utils

import "log"

func LogInfo(msg string) {
	log.Printf("[INFO]: %s", msg)
}

func LogInfoF(msg string, value string) {
	log.Printf("[INFO] \n message:%s \n error: %s \n", msg, value)
}

func LogError(msg string, err error) {
	log.Fatalf("[ERROR] \n message :%s \n error:%s", msg, err)
}
