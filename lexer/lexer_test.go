package lexer

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestGenerateTimeString(t *testing.T) {
	Convey("Given a time", t, func() {
		testTime := time.Date(2009, time.September, 10, 10, 35, 47, 0, time.UTC)
		testTimeString := "2009_09_10_10_35_47"

		Convey("The two string should be equal", func() {
			generatedTime := GenerateTimeString(testTime)

			So(generatedTime, ShouldEqual, testTimeString)
		})
	})
}

func TestReadFile(t *testing.T) {
	Convey("Given a file name", t, func() {
		tempFile, _ := ioutil.TempFile(os.TempDir(), "testFile")
		fileName := tempFile.Name()

		testString := "I am some data!"
		tempFile.WriteString(testString)

		Convey("If the file exists, it should return data", func() {
			data, err := ReadFile(fileName)
			So(string(data), ShouldEqual, testString)
			So(err, ShouldBeNil)
		})

		Convey("If the file does not exist, it should return error", func() {
			data, err := ReadFile("some_non_existent_file")
			So(data, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			os.Remove(fileName)
		})
	})
}

func TestListingFile(t *testing.T) {
	Convey("Given a listing file", t, func() {
		Convey("Add a line to it", func() {

		})
		Convey("Add an error to it", func() {

		})
		Convey("Save the listing file", func() {

		})
	})
}
