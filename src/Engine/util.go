package Engine

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"google.golang.org/protobuf/proto"
)

func GetLocPositionFromSource(sourceCode []string, start int, end int) *pb.LinesOfCode {

	var loc, cloc, lloc, blankLines int

	// Count lines of code
	loc = end - start + 1
	cloc = 0 //countComments(x)
	lloc = loc
	blankLines = 0

	// get blank lines (line breaks) and declaration line
	for i := start - 1; i < end; i++ {
		// trim it
		sourceCode[i] = strings.TrimSpace(sourceCode[i])

		if sourceCode[i] == "" {
			lloc--
			blankLines++
		}

		// if beginning of line is not a comment, it's a declaration line
		if strings.HasPrefix(sourceCode[i], "//") ||
			strings.HasPrefix(sourceCode[i], "/*") ||
			strings.HasPrefix(sourceCode[i], "*/") ||
			strings.HasPrefix(sourceCode[i], "*") ||
			strings.HasPrefix(sourceCode[i], "\"") ||
			strings.HasPrefix(sourceCode[i], "#") {
			// @todo issue here.
			// Please update it using the countComments() function
			lloc--
			cloc++
		}
	}

	linesOfCode := pb.LinesOfCode{}
	linesOfCode.LinesOfCode = int32(loc)
	linesOfCode.CommentLinesOfCode = int32(cloc)
	// lloc = loc - (clocl + blank lines + declaration line)
	lloc = loc - (cloc + blankLines + 2)
	linesOfCode.LogicalLinesOfCode = int32(lloc)

	return &linesOfCode
}

func DumpProtobuf(file *pb.File, binPath string) {
	out, err := proto.Marshal(file)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(binPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Fatal(err)
	}
}

// FactoryStmts returns a new instance of Stmts
func FactoryStmts() *pb.Stmts {

	stmts := &pb.Stmts{}
	stmts.StmtDecisionIf = []*pb.StmtDecisionIf{}
	stmts.StmtDecisionSwitch = []*pb.StmtDecisionSwitch{}
	stmts.StmtDecisionCase = []*pb.StmtDecisionCase{}
	stmts.StmtLoop = []*pb.StmtLoop{}
	stmts.StmtFunction = []*pb.StmtFunction{}
	stmts.StmtClass = []*pb.StmtClass{}

	return stmts
}

// Provides the hash of a file, in order to avoid to parse it twice
func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func GetClassesInFile(file *pb.File) []*pb.StmtClass {
	var classes []*pb.StmtClass
	for _, namespace := range file.Stmts.StmtNamespace {
		classes = append(classes, namespace.Stmts.StmtClass...)
	}
	classes = append(classes, file.Stmts.StmtClass...)
	return classes
}

func GetFunctionsInFile(file *pb.File) []*pb.StmtFunction {
	var functions []*pb.StmtFunction
	for _, namespace := range file.Stmts.StmtNamespace {
		functions = append(functions, namespace.Stmts.StmtFunction...)
	}
	classes := GetClassesInFile(file)
	for _, class := range classes {
		functions = append(functions, class.Stmts.StmtFunction...)
	}
	functions = append(functions, file.Stmts.StmtFunction...)
	return functions
}
