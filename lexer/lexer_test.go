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
		listingFile := new(ListingFile)
		fileName := GenerateTimeString(time.Now()) + "_listing_file.txt"

		Convey("Add a line to it", func() {
			newLine := "Some source code!"

			listingFile.AddLine(newLine)

			So(listingFile.buf.String(), ShouldEqual, "1: Some source code!\n")
			So(listingFile.counter, ShouldEqual, 1)
		})

		Convey("Add two lines to it", func() {
			newLine := "Some source code!"
			newLine2 := "Some more source code!"

			listingFile.AddLine(newLine)
			listingFile.AddLine(newLine2)

			So(listingFile.buf.String(), ShouldEqual, "1: Some source code!\n2: Some more source code!\n")
			So(listingFile.counter, ShouldEqual, 2)
		})

		Convey("Add an error to it", func() {
			someError := "This is wrong."

			listingFile.AddError(someError)

			So(listingFile.buf.String(), ShouldEqual, "LEXERR: This is wrong.\n")
			So(listingFile.counter, ShouldEqual, 0)
		})

		Convey("Add two errors to it", func() {
			someError := "This is wrong."
			someOtherError := "This is also wrong."

			listingFile.AddError(someError)
			listingFile.AddError(someOtherError)

			So(listingFile.buf.String(), ShouldEqual, "LEXERR: This is wrong.\nLEXERR: This is also wrong.\n")
			So(listingFile.counter, ShouldEqual, 0)
		})

		Convey("Add one line, one error, and another line to it", func() {
			someLine := "Some incorrect source code."
			someError := "Error describing incorrect source code."
			someOtherLine := "And some more source code here."

			listingFile.AddLine(someLine)
			listingFile.AddError(someError)
			listingFile.AddLine(someOtherLine)

			So(listingFile.buf.String(), ShouldEqual,
				"1: Some incorrect source code.\n"+
					"LEXERR: Error describing incorrect source code.\n"+
					"2: And some more source code here.\n")
			So(listingFile.counter, ShouldEqual, 2)
		})

		Convey("Save the listing file", func() {
			listingFile.AddLine("Some source code!")
			listingFile.Save(fileName)
			data, err := ioutil.ReadFile(fileName)
			dataString := string(data)

			So(dataString, ShouldEqual, "1: Some source code!\n")
			So(err, ShouldBeNil)
		})

		Reset(func() {
			os.Remove(fileName)
		})
	})
}
