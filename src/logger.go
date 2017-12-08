package main

import  (
	"os"
	"io"
	olog "github.com/sirupsen/logrus"
)
var log *olog.Logger 

var level map[string]olog.Level = map[string]olog.Level{
	"info" : olog.InfoLevel,
	"warn" : olog.WarnLevel,
	"debug" : olog.DebugLevel,
	"error" : olog.ErrorLevel,
}

func init() {
	log = GetLog()
}

func GetLog() *olog.Logger {
	var logdst io.Writer
	if Config.Logpath != "" {
		logdst, _ = os.OpenFile(Config.Logpath,os.O_APPEND|os.O_RDWR|os.O_CREATE,0644)
	} else {
		logdst = os.Stdout
	}
	
	loglevel := Config.Loglevel
	log = olog.New()
	log.Out = logdst
	if _,exist := level[loglevel];exist {
		log.SetLevel(level[loglevel])
	}else{
		panic("wrong default log level")
	}
	return log
}
