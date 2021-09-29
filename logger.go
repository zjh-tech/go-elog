package elog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	out           io.Writer //os.Stderr -> File
	level         int
	fileDirPrefix string
	buffEvents    chan *LogEvent
	exitChan      chan struct{}
	lastTime      time.Time
	file          *os.File
	buf           bytes.Buffer
	callDepth     int
	closeFlag     bool
	wg            sync.WaitGroup
}

//level: debug 0 info 1 warn 2 error 3
func NewLogger(fileDirPrefix string, level int) *Logger {
	logger := &Logger{
		out:           os.Stderr,
		level:         level,
		fileDirPrefix: fileDirPrefix,
		buffEvents:    make(chan *LogEvent, LogBuffEventSize),
		exitChan:      make(chan struct{}),
		callDepth:     LogCallDepth,
		closeFlag:     false,
	}

	return logger
}

func (l *Logger) Init() {
	l.startWriterGoroutine()
}

func (l *Logger) UnInit() {
	l.close()
	l.wg.Wait()
}

func (l *Logger) Debug(v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprint(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.addEvent(LogError, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.addEvent(LogError, fmt.Sprintf(format, v...))
}

func (l *Logger) startWriterGoroutine() {
	fmt.Printf("Log Goroutine Start \n")
	l.wg.Add(1)

	go func() {
		defer func() {
			fmt.Printf(" Log Goroutine Exit \n")
			l.wg.Done()
		}()

		exit := false
		for {
			if exit && len(l.buffEvents) == 0 {
				//ensure all log write file
				l.out.Write(l.buf.Bytes())
				return
			}
			select {
			case evt := <-l.buffEvents:
				l.outPut(evt.level, evt.content, evt.file, evt.line)
				GLogEventPool.Put(evt)
			case <-l.exitChan:
				l.exitChan = nil
				exit = true
			}
		}
	}()
}

func (l *Logger) close() {
	l.closeFlag = true
	close(l.exitChan)
}

func (l *Logger) addEvent(level int, content string) {
	if l.closeFlag {
		return
	}

	if l.level > level {
		return
	}

	_, file, line, _ := runtime.Caller(l.callDepth)
	index := strings.LastIndex(file, "/")
	partFileName := file
	if index != 0 {
		fileLen := len(file)
		partFileName = file[index+1 : fileLen]
	}

	event := GLogEventPool.Get().(*LogEvent)
	event.level = level
	event.content = content
	event.file = partFileName
	event.line = line

	l.buffEvents <- event
}

func (l *Logger) outPut(level int, content string, file string, line int) {
	//time zone
	now := time.Now()
	l.ensureFileExist(now)
	l.lastTime = now

	//example: 2021-08-18 19:55:21 [INFO] [benchmark.go192] Benchmark Info
	l.buf.WriteString(now.Format("2006-01-02 15:04:05"))
	l.buf.WriteByte(' ')
	l.buf.WriteString(loglevels[level])
	l.buf.WriteString(" [")
	l.buf.WriteString(file)
	l.buf.WriteString(strconv.Itoa(line))
	l.buf.WriteString("] ")
	l.buf.WriteString(content)

	//add \n
	if len(content) > 0 && content[len(content)-1] != '\n' {
		l.buf.WriteByte('\n')
	}

	l.out.Write(l.buf.Bytes())
	l.outPutConsole(l.buf.String())
	l.buf.Reset()
}

func (l *Logger) ensureFileExist(now time.Time) {
	if checkDiffDate(now, l.lastTime) {
		year, month, day := now.Date()
		dir := fmt.Sprintf("%d-%02d-%02d", year, month, day)
		filename := fmt.Sprintf("%d%02d%02d_%02d.log", year, month, day, now.Hour())
		l.createLogFile(dir, filename)
	}
}

func (l *Logger) createLogFile(dir string, filename string) {
	var file *os.File
	fullDir := l.fileDirPrefix + "/" + dir
	_ = createMutiDir(fullDir)
	fullFilePath := fullDir + "/" + filename
	if isExistPath(fullFilePath) {
		file, _ = os.OpenFile(fullFilePath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		file, _ = os.OpenFile(fullFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}

	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
		l.out = os.Stderr
	}

	l.file = file
	l.out = file
}

func isExistPath(path string) bool {
	var err error
	if _, err = os.Stat(path); err == nil {
		return true
	} else if os.IsExist(err) {
		return true
	}

	return false
}

func createMutiDir(filePath string) error {
	if !isExistPath(filePath) {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func checkDiffDate(now time.Time, last time.Time) bool {
	year, month, day := now.Date()
	hour, _, _ := now.Clock()

	yearl, monthl, dayl := last.Date()
	hourl, _, _ := last.Clock()

	return (year != yearl) || (month != monthl) || (day != dayl) || (hour != hourl)
}
