package mr

import "log"
import "os"

func PrintDebug(args interface{}) {
	if shouldPrintMesg() {
		log.Print(args)
 }
}

func PrintDebugf(mesg string, args ...interface{}) {
	if shouldPrintMesg() {
		log.Printf(mesg, args...)
 }
}

func shouldPrintMesg() bool {
	debug, exists := os.LookupEnv("DEBUG")

	return exists && debug == "true"
}