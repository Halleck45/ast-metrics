package Analyzer

type Comparator struct {
}

type Comparaison struct {
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
}

func NewComparator() *Comparator {
	return &Comparator{}
}

func (c *Comparator) Compare(first Aggregated, second Aggregated) Comparaison {
	comparaison := Comparaison{}

	// Compare the two projects
	comparaison.NbFiles = first.NbFiles - second.NbFiles
	comparaison.NbFunctions = first.NbFunctions - second.NbFunctions
	comparaison.NbClasses = first.NbClasses - second.NbClasses
	comparaison.NbClassesWithCode = first.NbClassesWithCode - second.NbClassesWithCode
	comparaison.NbMethods = first.NbMethods - second.NbMethods
	comparaison.Loc = first.Loc - second.Loc
	comparaison.Cloc = first.Cloc - second.Cloc
	comparaison.Lloc = first.Lloc - second.Lloc
	comparaison.AverageMethodsPerClass = first.AverageMethodsPerClass - second.AverageMethodsPerClass
	comparaison.AverageLocPerMethod = first.AverageLocPerMethod - second.AverageLocPerMethod
	comparaison.AverageLlocPerMethod = first.AverageLlocPerMethod - second.AverageLlocPerMethod
	comparaison.AverageClocPerMethod = first.AverageClocPerMethod - second.AverageClocPerMethod
	comparaison.AverageCyclomaticComplexityPerMethod = first.AverageCyclomaticComplexityPerMethod - second.AverageCyclomaticComplexityPerMethod
	comparaison.AverageCyclomaticComplexityPerClass = first.AverageCyclomaticComplexityPerClass - second.AverageCyclomaticComplexityPerClass
	comparaison.MinCyclomaticComplexity = first.MinCyclomaticComplexity - second.MinCyclomaticComplexity
	comparaison.MaxCyclomaticComplexity = first.MaxCyclomaticComplexity - second.MaxCyclomaticComplexity
	comparaison.AverageHalsteadDifficulty = first.AverageHalsteadDifficulty - second.AverageHalsteadDifficulty
	comparaison.AverageHalsteadEffort = first.AverageHalsteadEffort - second.AverageHalsteadEffort
	comparaison.AverageHalsteadVolume = first.AverageHalsteadVolume - second.AverageHalsteadVolume
	comparaison.AverageHalsteadTime = first.AverageHalsteadTime - second.AverageHalsteadTime
	comparaison.AverageHalsteadBugs = first.AverageHalsteadBugs - second.AverageHalsteadBugs
	comparaison.SumHalsteadDifficulty = first.SumHalsteadDifficulty - second.SumHalsteadDifficulty
	comparaison.SumHalsteadEffort = first.SumHalsteadEffort - second.SumHalsteadEffort
	comparaison.SumHalsteadVolume = first.SumHalsteadVolume - second.SumHalsteadVolume
	comparaison.SumHalsteadTime = first.SumHalsteadTime - second.SumHalsteadTime
	comparaison.SumHalsteadBugs = first.SumHalsteadBugs - second.SumHalsteadBugs
	comparaison.AverageMI = first.AverageMI - second.AverageMI
	comparaison.AverageMIwoc = first.AverageMIwoc - second.AverageMIwoc
	comparaison.AverageMIcw = first.AverageMIcw - second.AverageMIcw
	comparaison.AverageMIPerMethod = first.AverageMIPerMethod - second.AverageMIPerMethod
	comparaison.AverageMIwocPerMethod = first.AverageMIwocPerMethod - second.AverageMIwocPerMethod
	comparaison.AverageMIcwPerMethod = first.AverageMIcwPerMethod - second.AverageMIcwPerMethod
	comparaison.AverageAfferentCoupling = first.AverageAfferentCoupling - second.AverageAfferentCoupling
	comparaison.AverageEfferentCoupling = first.AverageEfferentCoupling - second.AverageEfferentCoupling
	comparaison.AverageInstability = first.AverageInstability - second.AverageInstability
	comparaison.CommitCountForPeriod = first.CommitCountForPeriod - second.CommitCountForPeriod
	comparaison.CommittedFilesCountForPeriod = first.CommittedFilesCountForPeriod - second.CommittedFilesCountForPeriod
	comparaison.BusFactor = first.BusFactor - second.BusFactor

	return comparaison
}
