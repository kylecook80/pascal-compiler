package main

/* Proj 1 */
// import (
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"os"
// 	"strconv"
// 	"time"
// )

// import scan "compiler/scanner"
// import _ "compiler/parser"
// import . "compiler/util"

/* End Proj 1 */

/* Proj 2 */
import (
	"fmt"
	_ "io"
	_ "io/ioutil"
	"os"
	_ "strconv"
	_ "time"
)

import scan "compiler/scanner"
import parse "compiler/parser"
import _ "compiler/util"

/* End Proj 2 */

func main() {
	// Get the arguments passed to the compiler
	args := os.Args
	file := args[1]

	if len(args) > 1 {
		scanner := scan.NewScanner()
		scanner.ReadReservedFile("scanner/reserved_words.list")
		scanner.ReadSourceFile(file)

		/* Proj 1 */
		// listing := NewListingFile()
		// tokenFile := []byte{}
		// source := ReadFile(file)

		// tok, err := scanner.NextToken()
		// for err != io.EOF {
		// 	tok, err = scanner.NextToken()

		// 	if err != nil {
		// 		listing.AddError(err.Error())
		// 	} else {
		// 		line := scanner.CurrentLineNumber() + 1
		// 		if tok.Type() != scan.WS {
		// 			tokenFile = append(tokenFile, []byte(strconv.Itoa(line)+": "+tok.String()+"\n")...)
		// 		}

		// 		if tok.Type() == scan.EOF {
		// 			break
		// 		}
		// 	}

		// 	if scanner.CurrentLineNumber() >= listing.LineCount() {
		// 		listing.AddLine(source.ReadLine(scanner.CurrentLineNumber()))
		// 	}
		// }

		// ioutil.WriteFile(GenerateTimeString(time.Now())+"_token_file.txt", tokenFile, 0644)
		// listing.Save()
		/* End Proj 1 */

		/* Proj 2 */
		parser := parse.NewParser(scanner)
		parser.Begin(file)
		/* End Proj 2 */
	} else {
		fmt.Println("Please specify a file name.")
	}
}
