package Analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	git2go "github.com/libgit2/git2go/v34"
)

type GitAnalyzer struct {
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
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
					Count: 0,
				}
			}
			filesByPath[relativePath] = file
		}

		// Get file history
		commits, err := repo.Walk()
		if err != nil {
			continue
		}

		commits.PushHead()
		commits.Sorting(git2go.SortTime)

		commits.Iterate(func(commit *git2go.Commit) bool {

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

			// Get diff
			diff, err := repo.DiffTreeToTree(parentTree, commitTree, nil)
			if err != nil {
				return true
			}

			// Get diff delta
			diff.ForEach(func(delta git2go.DiffDelta, progress float64) (git2go.DiffForEachHunkCallback, error) {

				// Ignore deleted files
				if delta.Status == git2go.DeltaDeleted {
					return nil, nil
				}

				// Ignore files that are not in the list
				if _, ok := filesByPath[delta.NewFile.Path]; !ok {
					return nil, nil
				}

				// Get file
				file := filesByPath[delta.NewFile.Path]

				// Increment commits count
				file.Commits.Count++

				return nil, nil
			}, git2go.DiffDetailFiles)
			return true
		})
	}
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