package Report

// The report structure serves to maintain a standardized
// report format.
// This class should not be exposed to other packages, as
// it is specific to JSON formatting.
type report struct {
	ConcernedFiles []file `json:"concernedFiles,omitempty"`
	// hashmap of classes, just with the qualified name, used for afferent coupling calculation
	ClassesAfferentCoupling              map[string]int            `json:"classesAfferentCoupling,omitempty"`
	NbFiles                              int                       `json:"numberFiles,omitempty"`
	NbFunctions                          int                       `json:"numberFunctions,omitempty"`
	NbClasses                            int                       `json:"numberClasses,omitempty"`
	NbClassesWithCode                    int                       `json:"numberClassesWithCode,omitempty"`
	NbMethods                            int                       `json:"numberMethods,omitempty"`
	Loc                                  int                       `json:"loc,omitempty"`
	Cloc                                 int                       `json:"cloc,omitempty"`
	Lloc                                 int                       `json:"lloc,omitempty"`
	AverageMethodsPerClass               float32                   `json:"averageMethodsPerClass,omitempty"`
	AverageLocPerMethod                  float32                   `json:"averageLocPerMethod,omitempty"`
	AverageLlocPerMethod                 float32                   `json:"averageLlocPerMethod,omitempty"`
	AverageClocPerMethod                 float32                   `json:"averageClocPerMethod,omitempty"`
	AverageCyclomaticComplexityPerMethod float32                   `json:"averageCyclomaticComplexityPerMethod,omitempty"`
	AverageCyclomaticComplexityPerClass  float32                   `json:"averageCyclomaticComplexityPerClass,omitempty"`
	MinCyclomaticComplexity              int                       `json:"minCyclomaticComplexity,omitempty"`
	MaxCyclomaticComplexity              int                       `json:"maxCyclomaticComplexity,omitempty"`
	AverageHalsteadDifficulty            float32                   `json:"averageHalsteadDifficulty,omitempty"`
	AverageHalsteadEffort                float32                   `json:"averageHalsteadEffort,omitempty"`
	AverageHalsteadVolume                float32                   `json:"averageHalsteadVolume,omitempty"`
	AverageHalsteadTime                  float32                   `json:"averageHalsteadTime,omitempty"`
	AverageHalsteadBugs                  float32                   `json:"averageHalsteadBugs,omitempty"`
	SumHalsteadDifficulty                float32                   `json:"sumHalsteadDifficulty,omitempty"`
	SumHalsteadEffort                    float32                   `json:"sumHalsteadEffort,omitempty"`
	SumHalsteadVolume                    float32                   `json:"sumHalsteadVolume,omitempty"`
	SumHalsteadTime                      float32                   `json:"sumHalsteadTime,omitempty"`
	SumHalsteadBugs                      float32                   `json:"sumHalsteadBugs,omitempty"`
	AverageMI                            float32                   `json:"averageMI,omitempty"`
	AverageMIwoc                         float32                   `json:"averageMIwoc,omitempty"`
	AverageMIcw                          float32                   `json:"averageMIcw,omitempty"`
	AverageMIPerMethod                   float32                   `json:"averageMIPerMethod,omitempty"`
	AverageMIwocPerMethod                float32                   `json:"averageMIwocPerMethod,omitempty"`
	AverageMIcwPerMethod                 float32                   `json:"averageMIcwPerMethod,omitempty"`
	AverageAfferentCoupling              float32                   `json:"averageAfferentCoupling,omitempty"`
	AverageEfferentCoupling              float32                   `json:"averageEfferentCoupling,omitempty"`
	AverageInstability                   float32                   `json:"averageInstability,omitempty"`
	CommitCountForPeriod                 int                       `json:"commitCountForPeriod,omitempty"`
	CommittedFilesCountForPeriod         int                       `json:"committedFilesCountForPeriod,omitempty"` // for example if one commit concerns 10 files, it will be 10
	BusFactor                            int                       `json:"busFactor,omitempty"`
	TopCommitters                        []contributor             `json:"topCommitters,omitempty"`
	GitAnalysis                          []gitAnalysis             `json:"gitAnalysis,omitempty"`
	PackageRelations                     map[string]map[string]int `json:"packageRelations,omitempty"` // counter of dependencies. Ex: A -> B -> 2
}

type contributor struct {
	Name  string `json:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

type gitAnalysis struct {
	ProgrammingLanguage     string
	ReportRootDir           string
	CountCommits            int
	CountCommiters          int
	CountCommitsForLanguage int
	CountCommitsIgnored     int
}

type file struct {
	Path            string          `json:"path,omitempty"`
	Complexity      complexity      `json:"complexity,omitempty"`
	Volume          volume          `json:"volume,omitempty"`
	Maintainability maintainability `json:"maintainability,omitempty"`
	Risk            risk            `json:"risk,omitempty"`
	Coupling        coupling        `json:"coupling,omitempty"`
}

type complexity struct {
	Cyclomatic int32 `json:"cyclomatic,omitempty"`
}

type risk struct {
	Score float32 `json:"score,omitempty"` // score of risk. Lower is better
}

type coupling struct {
	Afferent    int32   `json:"afferent,omitempty"`    // number of classes that depends on this class
	Efferent    int32   `json:"efferent,omitempty"`    // number of classes that this class depends on
	Instability float32 `json:"instability,omitempty"` // instability of the class
}

type maintainability struct {
	MaintainabilityIndex                float32 `json:"maintainabilityIndex,omitempty"`
	MaintainabilityIndexWithoutComments float32 `json:"maintainabilityIndexWithoutComments,omitempty"`
	CommentWeight                       float32 `json:"commentWeight,omitempty"`
}

type volume struct {
	Loc                     int32   `json:"loc,omitempty"`
	Lloc                    int32   `json:"lloc,omitempty"`
	Cloc                    int32   `json:"cloc,omitempty"`
	HalsteadVocabulary      int32   `json:"halsteadVocabulary,omitempty"`
	HalsteadLength          int32   `json:"halsteadLength,omitempty"`
	HalsteadVolume          float32 `json:"halsteadVolume,omitempty"`
	HalsteadDifficulty      float32 `json:"halsteadDifficulty,omitempty"`
	HalsteadEffort          float32 `json:"halsteadEffort,omitempty"`
	HalsteadTime            float32 `json:"halsteadTime,omitempty"`
	HalsteadEstimatedLength float32 `json:"halsteadEstimatedLength,omitempty"`
}
