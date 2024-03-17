package Engine

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/elliotchance/orderedmap/v2"

	log "github.com/sirupsen/logrus"

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

		// if line exceeds source code length, skip it
		if i >= len(sourceCode) {
			continue
		}

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
		log.Error(err)
	}

	f, err := os.Create(binPath)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Error(err)
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
	if file.Stmts == nil {
		return classes
	}
	if file.Stmts.StmtNamespace != nil {
		for _, namespace := range file.Stmts.StmtNamespace {
			classes = append(classes, namespace.Stmts.StmtClass...)
		}
	}
	classes = append(classes, file.Stmts.StmtClass...)
	return classes
}

func GetFunctionsInFile(file *pb.File) []*pb.StmtFunction {
	var functions []*pb.StmtFunction
	if file.Stmts == nil {
		return functions
	}

	if file.Stmts.StmtNamespace != nil {
		for _, namespace := range file.Stmts.StmtNamespace {
			functions = append(functions, namespace.Stmts.StmtFunction...)
		}
	}
	classes := GetClassesInFile(file)
	for _, class := range classes {
		if class.Stmts == nil {
			continue
		}

		functions = append(functions, class.Stmts.StmtFunction...)
	}
	functions = append(functions, file.Stmts.StmtFunction...)
	return functions
}

// render as HTML
func HtmlChartLine(data *orderedmap.OrderedMap[string, float64], label string, id string) string {
	series := "["
	for _, key := range data.Keys() {
		value, _ := data.Get(key)
		series += "{ x: \"" + key + "\", y: " + fmt.Sprintf("%f", value) + "},"
	}
	series += "]"
	html := `
	<div id="` + id + `"></div>
	<script type="text/javascript">
var options = {
  colors: ["#1A56DB"],
  series: [
    {
      name: "` + label + `",
      color: "#1A56DB",
      data: ` + series + `,
    },
  ],
  chart: {
    type: "bar",
    height: "120px",
    fontFamily: "Inter, sans-serif",
    toolbar: {
      show: false,
    },
  },
  plotOptions: {
    bar: {
      horizontal: false,
      columnWidth: "70%",
      borderRadiusApplication: "end",
      borderRadius: 8,
    },
  },
  tooltip: {
    shared: true,
    intersect: false,
    style: {
      fontFamily: "Inter, sans-serif",
    },
  },
  states: {
    hover: {
      filter: {
        type: "darken",
        value: 1,
      },
    },
  },
  stroke: {
    show: true,
    width: 0,
    colors: ["transparent"],
  },
  grid: {
    show: false,
    strokeDashArray: 4,
    padding: {
      left: 2,
      right: 2,
      top: -14
    },
  },
  dataLabels: {
    enabled: false,
  },
  legend: {
    show: false,
  },
  xaxis: {
    floating: false,
    labels: {
      show: true,
      style: {
        fontFamily: "Inter, sans-serif",
        cssClass: 'text-xs font-normal fill-gray-500 dark:fill-gray-400'
      }
    },
    axisBorder: {
      show: false,
    },
    axisTicks: {
      show: false,
    },
  },
  yaxis: {
    show: false,
  },
  fill: {
    opacity: 1,
  },
}


if (document.getElementById("` + id + `") && typeof ApexCharts !== 'undefined') {
  const chart = new ApexCharts(document.getElementById("` + id + `"), options);
  chart.render();
}
</script>`
	return html
}

// render as HTML
func HtmlChartArea(data *orderedmap.OrderedMap[string, float64], label string, id string) string {

	values := "["
	keys := "["
	for _, key := range data.Keys() {
		value, _ := data.Get(key)
		values += fmt.Sprintf("%f", value) + ","
		keys += "\"" + key + "\","
	}
	values += "]"
	keys += "]"

	html := `
	<div id="` + id + `"></div>
	<script type="text/javascript">
	var options = {
		chart: {
		  height: "100%",
		  maxWidth: "100%",
		  type: "area",
		  fontFamily: "Inter, sans-serif",
		  dropShadow: {
			enabled: false,
		  },
		  toolbar: {
			show: false,
		  },
		},
		tooltip: {
		  enabled: true,
		  x: {
			show: false,
		  },
		},
		fill: {
		  type: "gradient",
		  gradient: {
			opacityFrom: 0.55,
			opacityTo: 0,
			shade: "#1C64F2",
			gradientToColors: ["#1C64F2"],
		  },
		},
		dataLabels: {
		  enabled: false,
		},
		stroke: {
		  width: 6,
		},
		grid: {
		  show: false,
		  strokeDashArray: 4,
		  padding: {
			left: 2,
			right: 2,
			top: 0
		  },
		},
		series: [
		  {
			name: "` + label + `",
			data: ` + values + `,
			color: "#1A56DB",
		  },
		],
		xaxis: {
		  categories: ` + keys + `,
		  labels: {
			show: false,
		  },
		  axisBorder: {
			show: false,
		  },
		  axisTicks: {
			show: false,
		  },
		},
		yaxis: {
		  show: false,
		},
	  }


if (document.getElementById("` + id + `") && typeof ApexCharts !== 'undefined') {
  const chart = new ApexCharts(document.getElementById("` + id + `"), options);
  chart.render();
}
</script>`
	return html
}
