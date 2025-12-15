package utils

import "log"

func LogInfo(msg string) {
	log.Printf("[INFO]: %s", msg)
}

func LogInfoF(msg string, value any) {
	log.Printf("[INFO] \n message:%s \n value: %v \n", msg, value)
}

func LogError(msg string, err error) {
	log.Printf("[ERROR] \n message: %s \n error: %v", msg, err)
}

func LogErrorf(msg string) {
	log.Printf("[ERROR] \n message: %s \n", msg)
}
