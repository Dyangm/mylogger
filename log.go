package mylogger

import (
	"encoding/json"
	"runtime"

	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

type MyJSONFormatter struct {
	Time  string `json:"time"`
	File  string `json:"file"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

type MyTextFormatter struct {
	Time  string
	Level string
	File  string
	Msg   string
}

func (f *MyJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{}, len(entry.Data)+3)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	jf := &(logrus.JSONFormatter{})
	bytes, _ := jf.Format(entry)

	json.Unmarshal(bytes, &f)
	lastIndex := strings.LastIndex(entry.Caller.File, `/`)
	lastIndex = strings.LastIndex(entry.Caller.File[:lastIndex], `/`)
	file := entry.Caller.File[lastIndex+1:] + fmt.Sprintf(":%d", entry.Caller.Line)
	f.Time = entry.Time.Format("2006-01-02 15:04:05.000")
	var serialized []byte
	var err error
	if len(data) > 0 {
		data["file"] = file
		data["time"] = f.Time
		data["msg"] = entry.Message
		data["level"] = entry.Level.String()

		serialized, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("FailEd to marshal fields to JSON, %v", err)
		}
	} else {
		f.File = file
		serialized, err = json.Marshal(f)
		if err != nil {
			return nil, fmt.Errorf("FailEd to marshal fields to JSON, %v", err)
		}
	}

	return append(serialized, '\n'), nil
}

func (myF *MyTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var keys []string = make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	b := &bytes.Buffer{}
	logrustF := &(logrus.TextFormatter{})
	if !logrustF.DisableSorting {
		sort.Strings(keys)
	}

	skip := 8
	if len(keys) > 0 {
		skip = 6
	}
	file, line := getSource(skip)
	myF.File = fmt.Sprintf("%s:%d", file, line)

	appendKeyValue(b, "time", entry.Time.Format("2006-01-02 15:04:05.000"))
	appendKeyValue(b, "address", myF.File)
	appendKeyValue(b, "level", entry.Level.String())
	if entry.Message != "" {
		appendKeyValue(b, "msg", entry.Message)
	}
	for _, key := range keys {
		if key == "msg" {
			continue
		}
		appendKeyValue(b, key, entry.Data[key])
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	appendValue(b, value)
}

func appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	if !needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func needsQuoting(text string) bool {
	if len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func getSource(num int) (string, int) {

	_, file, line, _ := runtime.Caller(num)
	last, arr := 0, file
	for i := 0; i < 2; i++ {
		last = strings.LastIndex(arr, "/")
		arr = arr[:last]
	}

	file = file[last+1:]
	return file, line
}
