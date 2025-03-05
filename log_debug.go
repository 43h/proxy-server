//go:build debug

package main

import (
	"log"
)

func LOGD(v ...interface{}) {
	if logLevel <= DEBUG {
		log.Println("[DEBUG] ", v)
	}
}

func LOGI(v ...interface{}) {
	if logLevel <= INFO {
		log.Println("[ INFO] ", v)
	}
}

func LOGE(v ...interface{}) {
	if logLevel <= ERROR {
		log.Println("[ERROR] ", v)
	}
}
