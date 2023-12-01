package Analyzer

import (
	"fmt"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

type GitAnalyzer struct {
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

func (gitAnalyzer *GitAnalyzer) Start(files []*pb.File) {

	// go-git have issues with file analysis
	// https://github.com/go-git/go-git/issues/811
	// https://github.com/go-git/go-git/issues/137
	// https://github.com/go-git/go-git/issues/534
	//
	// So we will use a workaround:

	// Create a map of all files, indexed by path of the .git folder
	// This is to avoid opening the same .git folder multiple times
	// associative array of files, but filename is the key in order to speed up the search

	gitRepoMap := make(map[string]map[string]bool)

	var repoPath string
	for _, file := range files {
		openOptions := &git.PlainOpenOptions{DetectDotGit: true}
		gitRepo, err := git.PlainOpenWithOptions(file.Path, openOptions)
		if err != nil {
			repoPath = "unversioned"
		} else {
			worktree, _ := gitRepo.Worktree()
			repoPath = worktree.Filesystem.Root() + string(os.PathSeparator)
		}

		// add file to the list of files for this repo
		if _, ok := gitRepoMap[repoPath]; !ok {
			gitRepoMap[repoPath] = make(map[string]bool)
		}

		gitRepoMap[repoPath][file.Path] = true
	}

	// Map of commits
	// map of string to pb.Commits
	commitsMap := make(map[string]*pb.Commits)

	// for each repo, open it and analyze all files
	for repoPath, filesOfRepo := range gitRepoMap {
		if repoPath == "unversioned" {
			continue
		}

		// open repo
		openOptions := &git.PlainOpenOptions{DetectDotGit: true}
		gitRepo, err := git.PlainOpenWithOptions(repoPath, openOptions)

		if err != nil {
			fmt.Println("Error opening repo: ", err)
			continue
		}

		// start from HEAD
		ref, _ := gitRepo.Head()

		// limit to 1 year
		since := time.Now().AddDate(-1, 0, 0)

		// Execute the git log command
		options := &git.LogOptions{
			PathFilter: func(path string) bool {
				// return true if path in the list of files
				fullpath := repoPath + path

				if _, ok := filesOfRepo[fullpath]; !ok {
					return false
				}

				return true
			},
			From: ref.Hash(),
			// only 1 year
			Since: &since,
			Order: git.LogOrderCommitterTime,
		}
		commits, _ := gitRepo.Log(options)
		defer commits.Close()

		commits.ForEach(func(commit *object.Commit) error {

			parent, _ := commit.Parent(0)
			if parent == nil {
				return nil
			}

			fmt.Println("Commit: ", commit.Hash.String())

			// get diff between parent and commit
			diff, _ := parent.Patch(commit)
			patches := diff.FilePatches()
			for _, patch := range patches {

				from, to := patch.Files()
				if from == nil || to == nil {
					continue
				}

				// get filename
				filename := from.Path()
				fullpath := repoPath + filename

				// if not a file we are interested (filesOfRepo) in, skip
				if _, ok := filesOfRepo[fullpath]; !ok {
					return nil
				}

				if _, ok := commitsMap[fullpath]; !ok {
					commitsMap[fullpath] = &pb.Commits{Count: 0}
				}

				fmt.Println("Commit: ", commit.Hash.String(), "File: ", fullpath)
				commitsMap[fullpath].Count++
				return nil
			}

			return nil
		})

		// for each file, add the number of commits
		for _, file := range files {
			if _, ok := commitsMap[file.Path]; ok {
				file.Commits = commitsMap[file.Path]
			}
		}
	}
}
