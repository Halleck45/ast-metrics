package Analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	git2go "github.com/libgit2/git2go/v34"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GitAnalyzer struct {
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
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

			for _, file := range files {
				relativePath := file.Path[len(repoRoot)+1:]
				cId, _ := commitTree.EntryByPath(relativePath)
				pId, _ := parentTree.EntryByPath(relativePath)

				if cId == nil {
					// deleted file
					continue
				}

				if pId == nil {
					// new file
					continue
				}

				// count commits
				if pId != nil && cId != nil {
					if cId.Id.String() != pId.Id.String() {

						if commit.ParentCount() > 1 {
							// Exclude merge commits
							return true
						}

						file.Commits.CountCommits++

						// append committer to collection
						committer := pb.CommitsHistory{}
						committer.Email = commit.Committer().Email
						committer.Name = commit.Committer().Name
						committer.Date = timestamppb.New(commit.Committer().When)
						file.Commits.Committers = append(file.Commits.Committers, &committer)

						// committers
						committersByFile[relativePath][commit.Committer().Email] = true
					}
				}
			}

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
