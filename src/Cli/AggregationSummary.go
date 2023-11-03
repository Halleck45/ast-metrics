package Cli

import (
    "strconv"
    "fmt"
    "github.com/halleck45/ast-metrics/src/Analyzer"
    "github.com/charmbracelet/glamour"
)

func AggregationSummary(aggregated Analyzer.Aggregated) {

   style := StyleTitle()
   fmt.Println(style.Render("Results overview"))

   var percentageCloc int = 0
   var percentageLloc int = 0
   if aggregated.Loc > 0 {
        percentageCloc = 100 * aggregated.Cloc / aggregated.Loc
        percentageLloc = 100 * aggregated.Lloc / aggregated.Loc
   }

   in := `*This code is composed from ` +
        strconv.Itoa(aggregated.Loc) + ` lines of code, ` +
        strconv.Itoa(aggregated.Cloc) + ` (` + ( strconv.Itoa(percentageCloc) )+ `%) comment lines of code and ` +
        strconv.Itoa(aggregated.Lloc) + ` (` + ( strconv.Itoa(percentageLloc) )+ `%) logical lines of code.*

   ## Complexity

   ### Cyclomatic complexity

   *Cyclomatic Complexity is a measure of the number of linearly independent paths through a program's source code.
   More you have paths, more your code is complex.*


   | Min | Max | Average per class | Average per method |
   | --- | --- | --- | --- |
   | ` +
        strconv.Itoa(aggregated.MinCyclomaticComplexity) +
        ` | ` + strconv.Itoa(aggregated.MaxCyclomaticComplexity) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageCyclomaticComplexityPerClass) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageCyclomaticComplexityPerMethod) +
        ` |

   ### Halstead metrics

   *Halstead metrics are software metrics introduced to empirically determine the complexity of a program.*

   | | Difficulty | Effort | Volume | Time |
   | --- | --- | --- | --- | --- |
    ` +
        ` | Total` +
        ` | ` + fmt.Sprintf("%.2f", aggregated.SumHalsteadDifficulty) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.SumHalsteadEffort) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.SumHalsteadVolume) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.SumHalsteadTime) +
        "\n | Average per class" +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageHalsteadDifficulty) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageHalsteadEffort) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageHalsteadVolume) +
        ` | ` + fmt.Sprintf("%.2f", aggregated.AverageHalsteadTime) +
        ` |

   ### Classes and methods

   | Classes | Methods | Average methods per class |
   | --- | --- | --- |
   | ` + strconv.Itoa(aggregated.NbClasses) + ` | ` + strconv.Itoa(aggregated.NbMethods) + ` | ` + fmt.Sprintf("%.2f", aggregated.AverageMethodsPerClass) + ` |

   ## Maintainability

   *Maintainability Index is a software metric which measures how maintainable (easy to support and change) the source code is.
   If you have a high MI (>85), your code is easy to maintain.*

   | Maintainability index | MI without comments | Comment weight |
   | --- | --- | --- |
   | ` + DecorateMaintainabilityIndex(int(aggregated.AverageMI)) + ` | ` + fmt.Sprintf("%.2f", aggregated.AverageMIwoc) + ` | ` + fmt.Sprintf("%.2f", aggregated.AverageMIcw) + ` |

   `
   out, _ := glamour.Render(in, "dark")
   fmt.Print(out)
}