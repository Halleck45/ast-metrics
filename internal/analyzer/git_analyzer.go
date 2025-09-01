package analyzer

import (
	"path/filepath"
	"strings"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/scm"
	log "github.com/sirupsen/logrus"
)

type GitAnalyzer struct {
	git scm.GitRepository
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

type gitLogOutput struct {
	lines []string
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
		repoRoot, err := scm.FindGitRoot(file.Path)
		if err != nil {
			continue
		}

		// Declare the short path of the file (from repository root)
		// ex: /var/www/foo/bar.go -> foo/bar.go
		file.ShortPath = strings.TrimPrefix(file.Path, repoRoot)
		file.ShortPath = strings.TrimPrefix(file.ShortPath, "/")

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

		gitObject, err := scm.NewGitRepositoryFromPath(repoRoot)
		if err != nil {
			log.Debug("Not a valid git repository: ", repoRoot)
			continue
		}

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
		filesByPathInRepository := make(map[string]*pb.File)

		// Map of committers by file
		committersByFile := make(map[string]map[string]bool)

		for _, file := range filesByGitRepo[repoRoot] {
			// Add file to filesByPathInRepository map
			absolutePath := file.Path
			if !filepath.IsAbs(file.Path) {
				absolutePath = filepath.Join(gitObject.Path, file.Path)
			}

			if _, ok := filesByPathInRepository[absolutePath]; !ok {
				filesByPathInRepository[absolutePath] = file
			}
			// Add file to committersByFile map
			committersByFile[absolutePath] = make(map[string]bool)
		}

		// Get all commits once
		commits, err := gitObject.ListAllCommitsSince("1.year")
		if err != nil {
			log.Error("Error: ", err)
			continue
		}

		// For each commit
		summary.CountCommits = len(commits)

		for _, commit := range commits {

			doesCommitConcernsObservedProgrammingLanguage := false

			// For each file in the commit
			for _, file := range commit.Files {

				// make file absolute
				file = filepath.Join(gitObject.Path, file)

				// Get the file in the map filesByPathInRepository
				// If the file is not in the map, continue
				if _, ok := filesByPathInRepository[file]; !ok {
					continue
				}

				// Historize commit
				pbCommit := &pb.Commit{
					Hash:   commit.Hash,
					Date:   int64(commit.Timestamp),
					Author: commit.Author,
				}

				filesByPathInRepository[file].Commits.Count++
				filesByPathInRepository[file].Commits.Commits = append(filesByPathInRepository[file].Commits.Commits, pbCommit)
				committersByFile[file][commit.Author] = true

				doesCommitConcernsObservedProgrammingLanguage = true
			}

			// increment commit count
			if doesCommitConcernsObservedProgrammingLanguage {
				// add committer to the map
				committersOnRepository[commit.Author] = true

				summary.CountCommitsForLanguage++
			}

			// @todo: to examine:
			// Note: we may consider having two metrics: committersOnRepositoryForLanguage and committersOnRepository
		}

		summary.CountCommitsIgnored = summary.CountCommits - summary.CountCommitsForLanguage

		// Count committers
		for file, committers := range committersByFile {
			filesByPathInRepository[file].Commits.CountCommiters = 0
			if filesByPathInRepository[file].Commits == nil {
				filesByPathInRepository[file].Commits.CountCommiters = int32(len(committers))
			}

			// creation commit is counted
			if filesByPathInRepository[file].Commits.Count == 0 {
				filesByPathInRepository[file].Commits.CountCommiters = 1
			}
		}

		summary.CountCommiters = len(committersOnRepository)
		summary.GitRepository = gitObject
		summaries = append(summaries, summary)
	}

	return summaries
}
