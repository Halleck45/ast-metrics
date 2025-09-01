package engine

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	storage "github.com/halleck45/ast-metrics/internal/storage"
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
	cfg *configuration.Configuration,
	progress *pterm.SpinnerPrinter,
	parse func(path string) (*pb.File, error),
	opts DumpOptions,
) {
	if len(files) == 0 {
		return
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = runtime.NumCPU()
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
