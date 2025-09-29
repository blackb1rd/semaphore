package main

import (
	"log/syslog"

	"github.com/semaphoreui/semaphore/cli/cmd"
	log "github.com/sirupsen/logrus"

	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

func initSyslogs() {
	hook, err := lSyslog.NewSyslogHook("udp", "123.mgukov.com:514", syslog.LOG_DEBUG, "semaphoreui")
	if err == nil {
		log.AddHook(hook)
	} else {
		log.WithError(err).Fatal("Failed to create syslog hook")
	}
}

func main() {
	initSyslogs()
	cmd.Execute()
}
