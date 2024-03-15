package Analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	log "github.com/sirupsen/logrus"
)

type GitAnalyzer struct {
}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
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

func (gitAnalyzer *GitAnalyzer) Start(files []*pb.File) {
	gitAnalyzer.CalculateCount(files)
}

func (gitAnalyzer *GitAnalyzer) CalculateCount(files []*pb.File) {

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

		// Check if repo is a git repository, using the shell command "git rev-parse --is-inside-work-tree"
		// If not, continue to the next repository
		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
		cmd.Dir = repoRoot
		err := cmd.Run()
		if err != nil {
			log.Debug("Not a git repository: ", repoRoot)
			continue
		}

		// Create a hash map of files, indexed by relative path
		// Will be useful to retrieve file by path
		filesByPath := make(map[string]*pb.File)
		for _, file := range files {
			relativePath := file.Path[len(repoRoot)+1:]

			if file.Commits == nil {
				file.Commits = &pb.Commits{
					Count: 1, // creation commit is counted
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

		// Get all commits since one year (only sha1)
		// git --no-pager log --format=%H --since="1 year ago"
		cmd = exec.Command("git", "--no-pager", "log", "--pretty=format:%H", "--since=1.year")
		cmd.Dir = repoRoot
		out, err := cmd.Output()
		if err != nil {
			log.Error("Error: ", err)
			continue
		}
		// split output by line
		commits := strings.Split(string(out), "\n")
		for _, commit := range commits {
			// list modified, added, deleted files
			cmd = exec.Command("git", "log", "--pretty=format:%an|%ct", "--name-only", "-n", "1", commit)
			cmd.Dir = repoRoot
			out, err := cmd.Output()
			if err != nil {
				log.Error("Error: ", err)
				continue
			}

			lines := strings.Split(string(out), "\n")
			// first line is author email
			details := lines[0]
			// explode details by |
			authorEmail := strings.Split(details, "|")[0]
			date := strings.Split(details, "|")[1]

			// next lines are file paths
			impactedFiles := lines[1:]

			for _, file := range impactedFiles {
				// if file is not in the map, continue
				if _, ok := filesByPath[file]; !ok {
					continue
				}

				timestamp, err := strconv.ParseInt(date, 10, 64)
				if err != nil {
					log.Error("Error: ", err)
					continue
				}

				// increment commit count
				filesByPath[file].Commits.Count++
				commit := &pb.Commit{
					Hash:   commit,
					Date:   int64(timestamp),
					Author: authorEmail,
				}
				if filesByPath[file].Commits.Commits == nil {
					filesByPath[file].Commits.Commits = make([]*pb.Commit, 0)
				}
				filesByPath[file].Commits.Commits = append(filesByPath[file].Commits.Commits, commit)

				// add committer to the map
				committersByFile[file][authorEmail] = true
			}
		}

		// Count committers
		for file, committers := range committersByFile {
			filesByPath[file].Commits.CountCommiters = int32(len(committers))
		}
	}
}
