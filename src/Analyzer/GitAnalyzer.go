package Analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	log "github.com/sirupsen/logrus"
)

type GitAnalyzer struct {
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

type gitLogOutput struct {
	lines []string
}

func findGitRoot(filePath string) (string, error) {
	// Parcourir les dossiers parent jusqu'à ce qu'un dossier .git soit trouvé
	for filePath != "" && filePath != "/" {

		checkedPath := filepath.Join(filePath, ".git")
		if _, err := os.Stat(checkedPath); err == nil {
			return filePath, nil
		}

		filePath = filepath.Dir(filePath)
	}

	return "", fmt.Errorf("no git repository found")
}

func (gitAnalyzer *GitAnalyzer) Start(files []*pb.File) []ResultOfGitAnalysis {
	return gitAnalyzer.CalculateCount(files)
}

func (gitAnalyzer *GitAnalyzer) CalculateCount(files []*pb.File) []ResultOfGitAnalysis {

	summaries := make([]ResultOfGitAnalysis, 0)

	// Map of files by git repository
	filesByGitRepo := make(map[string][]*pb.File)

	// Prepare maps
	for _, file := range files {

		// Search root of git repository
		repoRoot, err := findGitRoot(file.Path)
		if err != nil {
			continue
		}

		// Add file to filesByGitRepo map
		if _, ok := filesByGitRepo[repoRoot]; !ok {
			filesByGitRepo[repoRoot] = make([]*pb.File, 0)
		}
		filesByGitRepo[repoRoot] = append(filesByGitRepo[repoRoot], file)

		// Prepare structures
		if file.Commits == nil {
			file.Commits = &pb.Commits{
				Count: 0,
			}
		}
		if file.Commits.Commits == nil {
			file.Commits.Commits = make([]*pb.Commit, 0)
		}
	}

	// For each git repository
	for repoRoot, _ := range filesByGitRepo {

		// Prepare result
		summary := ResultOfGitAnalysis{
			ReportRootDir:           repoRoot,
			CountCommits:            0,
			CountCommiters:          0,
			CountCommitsForLanguage: 0,
			CountCommitsIgnored:     0,
		}

		// Map of committers
		committersOnRepository := make(map[string]bool)

		// Map of files, by relative path
		filesByRelativePathInRepository := make(map[string]*pb.File)

		// Map of committers by file
		committersByFile := make(map[string]map[string]bool)

		for _, file := range filesByGitRepo[repoRoot] {
			// Add file to filesByRelativePathInRepository map
			relativePath := file.Path[len(repoRoot)+1:]
			if _, ok := filesByRelativePathInRepository[relativePath]; !ok {
				filesByRelativePathInRepository[relativePath] = file
			}

			// Add file to committersByFile map
			committersByFile[relativePath] = make(map[string]bool)
		}

		// Check if repo is a git repository, using the shell command "git rev-parse --is-inside-work-tree"
		// If not, continue to the next repository
		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
		cmd.Dir = repoRoot
		err := cmd.Run()
		if err != nil {
			log.Debug("Not a git repository: ", repoRoot)
			continue
		}

		// Get all commits since one year (only sha1)
		cmd = exec.Command("git", "--no-pager", "log", "--pretty=format:%H", "--since=1.year")
		cmd.Dir = repoRoot
		out, err := cmd.Output()
		if err != nil {
			log.Error("Error: ", err)
			continue
		}

		// Run git log on each sha1, in parallel
		// Wait for end of all goroutines
		var wg sync.WaitGroup

		commits := strings.Split(string(out), "\n")
		outputOfGitLog := make(chan []string, len(commits))
		defer close(outputOfGitLog)

		for _, commit := range commits {
			wg.Add(1)

			go func(commit string) {
				// Run git log on this sha1, in sub routine
				defer wg.Done()
				cmd := exec.Command("git", "log", "--pretty=format:%h|%an|%ct", "--name-only", "-n", "1", commit)
				cmd.Dir = repoRoot
				out, err := cmd.Output()

				if err != nil {
					log.Error("Error: ", err)
					outputOfGitLog <- strings.Split(string("ERROR"), "\n") // cannot be nil, because we need to escape channel
					return
				}

				outputOfGitLog <- strings.Split(string(out), "\n")
			}(commit)
		}

		// Wait for all git log to finish
		wg.Wait()

		// convert outputOfGitLog to slice
		results := make([][]string, 0, len(commits))
		for i := 0; i < len(commits); i++ {
			results = append(results, <-outputOfGitLog)
		}

		// For each git log output
		for _, lines := range results {

			// if error, continue
			if len(lines) == 1 && lines[0] == "ERROR" {
				continue
			}

			// first line is author email
			details := lines[0]

			// explode details by |
			sha1 := strings.Split(details, "|")[0]
			authorEmail := strings.Split(details, "|")[1]
			date := strings.Split(details, "|")[2]

			// next lines are file paths
			impactedFiles := lines[1:]

			nbFilesNotConcerned := 0

			for _, file := range impactedFiles {

				// if file is not in the map, continue
				if _, ok := filesByRelativePathInRepository[file]; !ok {
					// This case is normal, and occurs when a file is ignored
					// or not in the list of files to analyze
					nbFilesNotConcerned++
					continue
				}

				timestamp, err := strconv.ParseInt(date, 10, 64)
				if err != nil {
					log.Error("Error parsing data: ", err)
					continue
				}

				// increment commit count
				commit := &pb.Commit{
					Hash:   sha1,
					Date:   int64(timestamp),
					Author: authorEmail,
				}
				filesByRelativePathInRepository[file].Commits.Count++
				filesByRelativePathInRepository[file].Commits.Commits = append(filesByRelativePathInRepository[file].Commits.Commits, commit)

				// add committer to the map
				committersByFile[file][authorEmail] = true
				committersOnRepository[authorEmail] = true
			}

			summary.CountCommits++
			if nbFilesNotConcerned == len(impactedFiles) {
				summary.CountCommitsIgnored++
			} else {
				summary.CountCommitsForLanguage++
			}
		}

		// Count committers
		for file, committers := range committersByFile {
			filesByRelativePathInRepository[file].Commits.CountCommiters = 0
			if filesByRelativePathInRepository[file].Commits == nil {
				filesByRelativePathInRepository[file].Commits.CountCommiters = int32(len(committers))
			}

			// creation commit is counted
			if filesByRelativePathInRepository[file].Commits.Count == 0 {
				filesByRelativePathInRepository[file].Commits.CountCommiters = 1
			}
		}

		summary.CountCommiters = len(committersOnRepository)
		summaries = append(summaries, summary)
	}

	return summaries
}
