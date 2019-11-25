package log

import (
	"fmt"
	"os"
	"spider/utils"
	"github.com/a6910438/go-logger"
)

func Init(path, file string, level int) error {
	if file != "" {
		logPath := utils.AbsPath(path)
		logFile := file
		if !utils.PathExist(logPath) {
			os.MkdirAll(logPath, os.ModePerm)
		}
		err := logger.InitFileLog(logPath, logFile, level, true, false)
		if err != nil {
			fmt.Println("logger init error err=", err)
			return err
		}
	}
	logger.InitStdOutput(true, level, true)

	fmt.Println("logger inited")
	return nil
}
