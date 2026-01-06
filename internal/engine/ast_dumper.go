package engine

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/halleck45/ast-metrics/internal/configuration"
	storage "github.com/halleck45/ast-metrics/internal/storage"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
)

// Global options for DumpFiles (set by AIDatasetCommand)
var (
	globalMaxFiles    int
	globalConcurrency int
)

// SetDumpOptions sets global options for DumpFiles
func SetDumpOptions(maxFiles, concurrency int) {
	globalMaxFiles = maxFiles
	globalConcurrency = concurrency
}

type DumpOptions struct {
	Concurrency  int
	MaxFiles     int    // Maximum number of files to process (0 = unlimited)
	Label        string // ex: "PHP", "Go", "Python"
	BeforeParse  func(path string)
	AfterParse   func(file *pb.File)
	ProgressText func(done, total int, path string) string
}

func DumpFiles(
	files []string,
	cfg *configuration.Configuration,
	progress *pterm.SpinnerPrinter,
	parse func(path string) (*pb.File, error),
	opts DumpOptions,
) {
	if len(files) == 0 {
		if opts.Label != "" {
			pterm.Warning.Printf("No %s files found to dump\n", opts.Label)
		} else {
			pterm.Warning.Println("No files found to dump")
		}
		return
	}

	originalCount := len(files)

	// Limit files if MaxFiles is set
	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		if opts.Label != "" {
			pterm.Warning.Printf("Limiting %s files from %d to %d\n", opts.Label, len(files), opts.MaxFiles)
		} else {
			pterm.Warning.Printf("Limiting files from %d to %d\n", len(files), opts.MaxFiles)
		}
		files = files[:opts.MaxFiles]
	}

	if opts.Label != "" {
		pterm.Info.Printf("Found %d %s file(s) to process", originalCount, opts.Label)
		if opts.MaxFiles > 0 && originalCount > opts.MaxFiles {
			pterm.Info.Printf(" (processing %d due to --max-files limit)", len(files))
		}
		pterm.Info.Println()
	} else {
		pterm.Info.Printf("Found %d file(s) to process", originalCount)
		if opts.MaxFiles > 0 && originalCount > opts.MaxFiles {
			pterm.Info.Printf(" (processing %d due to --max-files limit)", len(files))
		}
		pterm.Info.Println()
	}

	// Use global concurrency if not set in options
	if opts.Concurrency <= 0 {
		if globalConcurrency > 0 {
			opts.Concurrency = globalConcurrency
		} else {
			opts.Concurrency = runtime.NumCPU()
			// Reduce default concurrency to avoid memory issues with large datasets
			if opts.Concurrency > 4 {
				opts.Concurrency = 4
			}
		}
	}

	// Use global MaxFiles if not set in options
	if opts.MaxFiles <= 0 && globalMaxFiles > 0 {
		opts.MaxFiles = globalMaxFiles
	}

	var wg sync.WaitGroup
	jobs := make(chan string, opts.Concurrency)
	total := len(files)
	done := 0
	var mu sync.Mutex

	if opts.ProgressText == nil {
		opts.ProgressText = func(done, total int, _ string) string {
			if opts.Label == "" {
				return fmt.Sprintf("Dumping AST (%d/%d)", done, total)
			}
			return fmt.Sprintf("Dumping %s files (%d/%d)", opts.Label, done, total)
		}
	}

	worker := func() {
		for path := range jobs {
			func(path string) {
				defer wg.Done()

				if opts.ProgressText != nil && progress != nil {
					mu.Lock()
					done++
					progress.UpdateText(opts.ProgressText(done, total, path))
					mu.Unlock()
				}

				if opts.BeforeParse != nil {
					opts.BeforeParse(path)
				}

				hash, err := storage.GetFileHash(path)
				if err != nil {
					return
				}
				bin := cfg.Storage.AstDirectory() + string(os.PathSeparator) + hash + ".bin"
				if _, err := os.Stat(bin); err == nil {
					return // ok: Done() déjà garanti par defer
				}
				if file, err := parse(path); err == nil && file != nil {
					file.Checksum = hash
					_ = DumpProtobuf(file, bin)
					if opts.AfterParse != nil {
						opts.AfterParse(file)
					}
				}
			}(path)
		}
	}

	for i := 0; i < opts.Concurrency; i++ {
		go worker()
	}
	for _, f := range files {
		wg.Add(1)
		jobs <- f
	}
	close(jobs)
	wg.Wait()
	if progress != nil {
		progress.Info("AST dumped")
	}
}
