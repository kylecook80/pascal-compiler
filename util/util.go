package util

import (
	"bytes"
	"errors"
	_ "fmt"
	"io"
	"os"
	"strings"
	"time"
)

const BUFSIZE = 32

type Buffer struct {
	bytes.Buffer
}

func (buf *Buffer) Lines() []string {
	return strings.Split(buf.String(), "\n")
}

func (buf *Buffer) ReadAt(idx int) (byte, error) {
	if idx < len(buf.Bytes()) {
		return buf.Bytes()[idx], nil
	} else {
		return 0, errors.New("Cannot read past size of buffer.")
	}
}

func (buf *Buffer) ReadLine(idx int) string {
	return buf.Lines()[idx]
}

// GenerateTimeString takes a time and formats it as an underscored
// string, suitable for a filename.
func GenerateTimeString(t time.Time) string {
	formattedTime := t.Format("2006-01-2-15-04-05")
	underscoreTime := strings.Replace(formattedTime, "-", "_", -1)
	return underscoreTime
}

func ReadFile(file string) *Buffer {
	openFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer openFile.Close()

	fileBuf := new(Buffer)

	for {
		readBuf := make([]byte, BUFSIZE)
		n, err := openFile.Read(readBuf)

		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		fileBuf.Write(readBuf[0:n])
	}
	return fileBuf
}
