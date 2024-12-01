package Analyzer

import (
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type Comparator struct {
	ComparedBranch string
}

const (
	ADDED     = "added"
	DELETED   = "deleted"
	MODIFIED  = "modified"
	UNCHANGED = "unchanged"
)

type ChangedFile struct {
	Path            string
	Comparaison     Comparaison
	Status          string
	LatestVersion   *pb.File
	PreviousVersion *pb.File
	IsNegligible    bool
}
type Comparaison struct {
	ComparedBranch                       string
	NbFiles                              int
	NbFunctions                          int
	NbClasses                            int
	NbClassesWithCode                    int
	NbMethods                            int
	Loc                                  int
	Cloc                                 int
	Lloc                                 int
	AverageMethodsPerClass               float64
	AverageLocPerMethod                  float64
	AverageLlocPerMethod                 float64
	AverageClocPerMethod                 float64
	AverageCyclomaticComplexityPerMethod float64
	AverageCyclomaticComplexityPerClass  float64
	MinCyclomaticComplexity              int
	MaxCyclomaticComplexity              int
	AverageHalsteadDifficulty            float64
	AverageHalsteadEffort                float64
	AverageHalsteadVolume                float64
	AverageHalsteadTime                  float64
	AverageHalsteadBugs                  float64
	SumHalsteadDifficulty                float64
	SumHalsteadEffort                    float64
	SumHalsteadVolume                    float64
	SumHalsteadTime                      float64
	SumHalsteadBugs                      float64
	AverageMI                            float64
	AverageMIwoc                         float64
	AverageMIcw                          float64
	AverageMIPerMethod                   float64
	AverageMIwocPerMethod                float64
	AverageMIcwPerMethod                 float64
	AverageAfferentCoupling              float64
	AverageEfferentCoupling              float64
	AverageInstability                   float64
	CommitCountForPeriod                 int
	CommittedFilesCountForPeriod         int // for example if one commit concerns 10 files, it will be 10
	BusFactor                            int
	Risk                                 float64
	ChangedFiles                         []ChangedFile
	NbNewFiles                           int
	NbDeletedFiles                       int
	NbModifiedFiles                      int
	NbModifiedFilesIncludingNegligible   int
}

func NewComparator(comparedBranch string) *Comparator {
	return &Comparator{
		ComparedBranch: comparedBranch,
	}
}

func (c *Comparator) Compare(first Aggregated, second Aggregated) Comparaison {
	comparaison := Comparaison{
		ComparedBranch: c.ComparedBranch,
	}

	// Compare the two projects
	comparaison.NbFiles = first.NbFiles - second.NbFiles
	comparaison.NbFunctions = first.NbFunctions - second.NbFunctions
	comparaison.NbClasses = first.NbClasses - second.NbClasses
	comparaison.NbClassesWithCode = first.NbClassesWithCode - second.NbClassesWithCode
	comparaison.NbMethods = first.NbMethods - second.NbMethods
	comparaison.Loc = int(first.Loc.Sum - second.Loc.Sum)
	comparaison.Cloc = int(first.Cloc.Sum - second.Cloc.Sum)
	comparaison.Lloc = int(first.Lloc.Sum - second.Lloc.Sum)
	comparaison.AverageMethodsPerClass = first.MethodsPerClass.Avg - second.MethodsPerClass.Avg
	comparaison.AverageLocPerMethod = first.LocPerMethod.Avg - second.LocPerMethod.Avg
	comparaison.AverageLlocPerMethod = first.LlocPerMethod.Avg - second.LlocPerMethod.Avg
	comparaison.AverageClocPerMethod = first.ClocPerMethod.Avg - second.ClocPerMethod.Avg
	comparaison.AverageCyclomaticComplexityPerMethod = first.CyclomaticComplexityPerMethod.Avg - second.CyclomaticComplexityPerMethod.Avg
	comparaison.AverageCyclomaticComplexityPerClass = first.CyclomaticComplexityPerClass.Avg - second.CyclomaticComplexityPerClass.Avg
	comparaison.MinCyclomaticComplexity = int(first.CyclomaticComplexityPerMethod.Min - second.CyclomaticComplexityPerMethod.Min)
	comparaison.MaxCyclomaticComplexity = int(first.CyclomaticComplexityPerMethod.Max - second.CyclomaticComplexityPerMethod.Max)
	comparaison.AverageHalsteadDifficulty = first.HalsteadDifficulty.Avg - second.HalsteadDifficulty.Avg
	comparaison.AverageHalsteadEffort = first.HalsteadEffort.Avg - second.HalsteadEffort.Avg
	comparaison.AverageHalsteadVolume = first.HalsteadVolume.Avg - second.HalsteadVolume.Avg
	comparaison.AverageHalsteadTime = first.HalsteadTime.Avg - second.HalsteadTime.Avg
	comparaison.AverageHalsteadBugs = first.HalsteadBugs.Avg - second.HalsteadBugs.Avg
	comparaison.SumHalsteadDifficulty = first.HalsteadDifficulty.Sum - second.HalsteadDifficulty.Sum
	comparaison.SumHalsteadEffort = first.HalsteadEffort.Sum - second.HalsteadEffort.Sum
	comparaison.SumHalsteadVolume = first.HalsteadVolume.Sum - second.HalsteadVolume.Sum
	comparaison.SumHalsteadTime = first.HalsteadTime.Sum - second.HalsteadTime.Sum
	comparaison.SumHalsteadBugs = first.HalsteadBugs.Sum - second.HalsteadBugs.Sum
	comparaison.AverageMI = first.MaintainabilityIndex.Avg - second.MaintainabilityIndex.Avg
	comparaison.AverageMIwoc = first.MaintainabilityIndexWithoutComments.Avg - second.MaintainabilityIndexWithoutComments.Avg
	comparaison.AverageMIPerMethod = first.MaintainabilityPerMethod.Avg - second.MaintainabilityPerMethod.Avg
	comparaison.AverageMIwocPerMethod = first.MaintainabilityCommentWeightPerMethod.Avg - second.MaintainabilityCommentWeightPerMethod.Avg
	comparaison.AverageMIcwPerMethod = first.MaintainabilityCommentWeightPerMethod.Avg - second.MaintainabilityCommentWeightPerMethod.Avg
	comparaison.AverageAfferentCoupling = first.AfferentCoupling.Avg - second.AfferentCoupling.Avg
	comparaison.AverageEfferentCoupling = first.EfferentCoupling.Avg - second.EfferentCoupling.Avg
	comparaison.AverageInstability = first.Instability.Avg - second.Instability.Avg
	comparaison.CommitCountForPeriod = first.CommitCountForPeriod - second.CommitCountForPeriod
	comparaison.CommittedFilesCountForPeriod = first.CommittedFilesCountForPeriod - second.CommittedFilesCountForPeriod
	comparaison.BusFactor = first.BusFactor - second.BusFactor

	for _, file := range first.ConcernedFiles {

		change := ChangedFile{
			Path: file.Path,
			Comparaison: Comparaison{
				NbFunctions:                          0,
				NbClasses:                            0,
				NbClassesWithCode:                    0,
				NbMethods:                            0,
				Loc:                                  0,
				Cloc:                                 0,
				Lloc:                                 0,
				AverageMethodsPerClass:               0,
				AverageLocPerMethod:                  0,
				AverageLlocPerMethod:                 0,
				AverageClocPerMethod:                 0,
				AverageCyclomaticComplexityPerMethod: 0,
				AverageCyclomaticComplexityPerClass:  0,
				MinCyclomaticComplexity:              0,
				MaxCyclomaticComplexity:              0,
				AverageHalsteadDifficulty:            0,
				AverageHalsteadEffort:                0,
				AverageHalsteadVolume:                0,
				AverageHalsteadTime:                  0,
				AverageHalsteadBugs:                  0,
				SumHalsteadDifficulty:                0,
				SumHalsteadEffort:                    0,
				SumHalsteadVolume:                    0,
				SumHalsteadTime:                      0,
				SumHalsteadBugs:                      0,
				AverageMI:                            0,
				AverageMIwoc:                         0,
				AverageMIcw:                          0,
				AverageMIPerMethod:                   0,
				AverageMIwocPerMethod:                0,
				AverageMIcwPerMethod:                 0,
				AverageAfferentCoupling:              0,
				AverageEfferentCoupling:              0,
				AverageInstability:                   0,
				CommitCountForPeriod:                 0,
				CommittedFilesCountForPeriod:         0,
				BusFactor:                            0,
				Risk:                                 0,
			},
			Status:          ADDED,
			LatestVersion:   file,
			PreviousVersion: nil,
			IsNegligible:    false,
		}

		for _, file2 := range second.ConcernedFiles {

			if file.Path != file2.Path {
				continue
			}

			if file.Checksum == file2.Checksum {
				// already present, no change
				change.Status = UNCHANGED
			}

			// nb functions
			change.Comparaison.NbFunctions = 0
			before := 0
			after := 0
			if file.Stmts != nil && file.Stmts.StmtFunction != nil {
				before = len(file.Stmts.StmtFunction)
			}
			if file2.Stmts != nil && file2.Stmts.StmtFunction != nil {
				after = len(file2.Stmts.StmtFunction)
			}
			change.Comparaison.NbFunctions = before - after

			// nb classes
			change.Comparaison.NbClasses = 0
			before = 0
			after = 0
			if file.Stmts != nil && file.Stmts.StmtClass != nil {
				before = len(file.Stmts.StmtClass)
			}
			if file2.Stmts != nil && file2.Stmts.StmtClass != nil {
				after = len(file2.Stmts.StmtClass)
			}
			change.Comparaison.NbClasses = before - after

			// Loc, cloc...
			if file.LinesOfCode != nil || file2.LinesOfCode != nil {
				change.Comparaison.Loc = int(file.LinesOfCode.LinesOfCode) - int(file2.LinesOfCode.LinesOfCode)
				change.Comparaison.Cloc = int(file.LinesOfCode.CommentLinesOfCode) - int(file2.LinesOfCode.CommentLinesOfCode)
				change.Comparaison.Lloc = int(file.LinesOfCode.LogicalLinesOfCode) - int(file2.LinesOfCode.LogicalLinesOfCode)
			}

			if file.Stmts.Analyze != nil && file2.Stmts.Analyze != nil {

				// Cyclomatic complexity
				if file.Stmts.Analyze.Complexity != nil && file2.Stmts.Analyze.Complexity != nil {
					change.Comparaison.AverageCyclomaticComplexityPerMethod = float64(*file.Stmts.Analyze.Complexity.Cyclomatic) - float64(*file2.Stmts.Analyze.Complexity.Cyclomatic)
				}

				// Halstead
				if file.Stmts.Analyze.Volume != nil && file.Stmts.Analyze.Volume.HalsteadDifficulty != nil &&
					file2.Stmts.Analyze.Volume != nil && file2.Stmts.Analyze.Volume.HalsteadDifficulty != nil {
					change.Comparaison.AverageHalsteadDifficulty = *file.Stmts.Analyze.Volume.HalsteadDifficulty - *file2.Stmts.Analyze.Volume.HalsteadDifficulty
					change.Comparaison.AverageHalsteadEffort = *file.Stmts.Analyze.Volume.HalsteadEffort - *file2.Stmts.Analyze.Volume.HalsteadEffort
					change.Comparaison.AverageHalsteadVolume = *file.Stmts.Analyze.Volume.HalsteadVolume - *file2.Stmts.Analyze.Volume.HalsteadVolume
					change.Comparaison.AverageHalsteadTime = *file.Stmts.Analyze.Volume.HalsteadTime - *file2.Stmts.Analyze.Volume.HalsteadTime
				}

				// Maintainability index
				if file.Stmts.Analyze.Maintainability != nil && file2.Stmts.Analyze.Maintainability != nil && file.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil && file2.Stmts.Analyze.Maintainability.MaintainabilityIndex != nil {
					change.Comparaison.AverageMI = *file.Stmts.Analyze.Maintainability.MaintainabilityIndex - *file2.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
				if file.Stmts.Analyze.Maintainability != nil && file2.Stmts.Analyze.Maintainability != nil && file.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil && file2.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments != nil {
					change.Comparaison.AverageMIwoc = *file.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments - *file2.Stmts.Analyze.Maintainability.MaintainabilityIndexWithoutComments
				}

				// Coupling
				if file.Stmts.Analyze.Coupling != nil && file2.Stmts.Analyze.Coupling != nil {
					change.Comparaison.AverageAfferentCoupling = float64(file.Stmts.Analyze.Coupling.Afferent) - float64(file2.Stmts.Analyze.Coupling.Afferent)
					change.Comparaison.AverageEfferentCoupling = float64(file.Stmts.Analyze.Coupling.Efferent) - float64(file2.Stmts.Analyze.Coupling.Efferent)
					change.Comparaison.AverageInstability = file.Stmts.Analyze.Coupling.Instability - file2.Stmts.Analyze.Coupling.Instability
				}

				// Risk
				if file.Stmts.Analyze.Risk != nil && file2.Stmts.Analyze.Risk != nil {
					change.Comparaison.Risk = file.Stmts.Analyze.Risk.Score - file2.Stmts.Analyze.Risk.Score
					// check if not NaN
					if change.Comparaison.Risk != change.Comparaison.Risk {
						change.Comparaison.Risk = 0
					}
				}
			}

			// if changes concerns only white spaces, etc. we don't want to include it
			change.IsNegligible = change.Comparaison.NbFunctions == 0 &&
				change.Comparaison.NbClasses == 0 &&
				change.Comparaison.Loc == 0 &&
				change.Comparaison.Cloc == 0 &&
				change.Comparaison.Lloc == 0 &&
				change.Comparaison.AverageCyclomaticComplexityPerMethod == 0 &&
				change.Comparaison.AverageHalsteadDifficulty == 0 &&
				change.Comparaison.AverageHalsteadEffort == 0 &&
				change.Comparaison.AverageHalsteadVolume == 0 &&
				change.Comparaison.AverageHalsteadTime == 0 &&
				change.Comparaison.AverageMI == 0 &&
				change.Comparaison.AverageMIwoc == 0 &&
				change.Comparaison.AverageAfferentCoupling == 0 &&
				change.Comparaison.AverageEfferentCoupling == 0 &&
				change.Comparaison.AverageInstability == 0 &&
				change.Comparaison.Risk == 0

			change.Status = MODIFIED
			change.PreviousVersion = file2
			change.LatestVersion = file
			break
		}

		if change.PreviousVersion != nil {
			if change.PreviousVersion.Checksum == change.LatestVersion.Checksum {
				// include only changed files
				continue
			}
		}

		switch change.Status {
		case ADDED:
			comparaison.NbNewFiles++
		case DELETED:
			comparaison.NbDeletedFiles++
		case MODIFIED:
			if change.IsNegligible {
				comparaison.NbModifiedFilesIncludingNegligible++
			} else {
				comparaison.NbModifiedFiles++
			}
		}

		if change.IsNegligible {
			continue
		}

		if change.Status == UNCHANGED {
			continue
		}

		comparaison.ChangedFiles = append(comparaison.ChangedFiles, change)
	}

	return comparaison
}
