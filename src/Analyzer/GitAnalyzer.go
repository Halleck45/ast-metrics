package Analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	git2go "github.com/libgit2/git2go/v34"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GitAnalyzer struct {
	monthsToConsider int // number of months to consider
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{monthsToConsider: 1} // tmp. Change me to 12
}

func findGitRoot(filePath string) (string, error) {
	// Parcourir les dossiers parent jusqu'à ce qu'un dossier .git soit trouvé
	for filePath != "" {

		checkedPath := filepath.Join(filePath, ".git")
		if _, err := os.Stat(checkedPath); err == nil {
			return filePath, nil
		}

		filePath = filepath.Dir(filePath)
	}

	return "", fmt.Errorf("no git repository found")
}

func (gitAnalyzer *GitAnalyzer) Start(files []*pb.File) {

	// Map of files by git repository
	filesByGitRepo := make(map[string][]*pb.File)

	for _, file := range files {
		// Search root of git repository
		repoRoot, err := findGitRoot(file.Path)
		if err != nil {
			continue
		}

		// Add file to map
		if _, ok := filesByGitRepo[repoRoot]; !ok {
			filesByGitRepo[repoRoot] = make([]*pb.File, 0)
		}

		filesByGitRepo[repoRoot] = append(filesByGitRepo[repoRoot], file)
	}

	// For each git repository
	for repoRoot, files := range filesByGitRepo {
		// Open repo
		repo, err := git2go.OpenRepository(repoRoot)
		if err != nil {
			continue
		}

		// Create a hash map of files, indexed by relative path
		// Will be useful to retrieve file by path
		filesByPath := make(map[string]*pb.File)
		for _, file := range files {
			relativePath := file.Path[len(repoRoot)+1:]

			if file.Commits == nil {
				file.Commits = &pb.Commits{
					CountCommits: 1, // creation commit is counted
				}
			}
			filesByPath[relativePath] = file
		}

		// Create a hash map of committers name by file
		committersByFile := make(map[string]map[string]bool)
		for _, file := range files {
			relativePath := file.Path[len(repoRoot)+1:]
			committersByFile[relativePath] = make(map[string]bool)
		}

		// Get file history
		commits, err := repo.Walk()
		if err != nil {
			continue
		}

		commits.PushHead()
		commits.Sorting(git2go.SortTime)

		now := time.Now()
		commits.Iterate(func(commit *git2go.Commit) bool {

			// If commit older than monthsToConsider, skip it
			if commit.Committer().When.Before(now.AddDate(0, -gitAnalyzer.monthsToConsider, 0)) {
				return true
			}

			// get the list of impacted files by this commit
			commitTree, err := commit.Tree()
			if err != nil {
				return true
			}

			// Compare with parent commit
			parents := commit.ParentCount()
			if parents == 0 {
				return true
			}

			parentCommit := commit.Parent(0)
			parentTree, err := parentCommit.Tree()
			if err != nil {
				return true
			}

			diff, _ := repo.DiffTreeToTree(parentTree, commitTree, nil)
			diff.ForEach(func(file git2go.DiffDelta, _ float64) (git2go.DiffForEachHunkCallback, error) {
				//relativePath := file.NewFile.Path[len(repoRoot)+1:]
				relativePath := file.NewFile.Path
				if file.Status == git2go.DeltaDeleted {
					// deleted file
					return nil, nil
				}

				// ensure file is in the list of files
				if _, ok := filesByPath[relativePath]; !ok {
					return nil, nil
				}

				// count commits
				filesByPath[relativePath].Commits.CountCommits++

				// append committer to collection
				committer := pb.CommitsHistory{}
				committer.Email = commit.Committer().Email
				committer.Name = commit.Committer().Name
				committer.Date = timestamppb.New(commit.Committer().When)
				filesByPath[relativePath].Commits.Committers = append(filesByPath[relativePath].Commits.Committers, &committer)

				// committers
				committersByFile[relativePath][commit.Committer().Email] = true

				return nil, nil
			}, git2go.DiffDetailFiles)

			return true
		})

		// count committers
		for _, file := range files {
			relativePath := file.Path[len(repoRoot)+1:]
			file.Commits.CountCommitters = int32(len(committersByFile[relativePath]))

			// if 0
			if file.Commits.CountCommitters == 0 {
				file.Commits.CountCommitters = 1
			}
		}
	}
}
