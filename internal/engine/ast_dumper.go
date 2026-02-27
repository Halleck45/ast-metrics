package engine

import (
	"fmt"
	"runtime"
	"sync"

	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
)

type DumpOptions struct {
	Concurrency  int
	Label        string // ex: "PHP", "Go", "Python"
	BeforeParse  func(path string)
	AfterParse   func(file *pb.File)
	ProgressText func(done, total int, path string) string
}

func DumpFiles(
	files []string,
	progress *pterm.SpinnerPrinter,
	parse func(path string) (*pb.File, error),
	opts DumpOptions,
) []*pb.File {
	if len(files) == 0 {
		return nil
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	jobs := make(chan string, opts.Concurrency)
	total := len(files)
	done := 0
	var mu sync.Mutex

	results := make([]*pb.File, 0, total)

	if opts.ProgressText == nil {
		opts.ProgressText = func(done, total int, _ string) string {
			if opts.Label == "" {
				return fmt.Sprintf("Parsing AST (%d/%d)", done, total)
			}
			return fmt.Sprintf("Parsing %s files (%d/%d)", opts.Label, done, total)
		}
	}

	worker := func() {
		for path := range jobs {
			if opts.ProgressText != nil && progress != nil {
				mu.Lock()
				done++
				progress.UpdateText(opts.ProgressText(done, total, path))
				mu.Unlock()
			}

			if opts.BeforeParse != nil {
				opts.BeforeParse(path)
			}

			if file, err := parse(path); err == nil && file != nil {
				if opts.AfterParse != nil {
					opts.AfterParse(file)
				}
				mu.Lock()
				results = append(results, file)
				mu.Unlock()
			}
			wg.Done()
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
		progress.Info("AST parsed")
	}
	return results
}
