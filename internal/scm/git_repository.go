package scm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type GitRepository struct {
	Path          string
	InitialBranch string
}

func NewGitRepositoryFromPath(path string) (GitRepository, error) {
	repoRoot, err := FindGitRoot(path)
	if err != nil {
		return GitRepository{}, err
	}

	// Get the absolute path of the repository
	absolutePath, err := getAbsolutePath(repoRoot)
	if err != nil {
		return GitRepository{}, err
	}

	// Ensure the path is a valid git repository
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = absolutePath
	err = cmd.Run()
	if err != nil {
		return GitRepository{}, fmt.Errorf("path is not a git repository")
	}

	gitRepository := GitRepository{
		Path: absolutePath,
	}

	// Get the current branch
	currentBranch, err := gitRepository.GetCurrentBranch()
	if err != nil {
		return GitRepository{}, err
	}
	gitRepository.InitialBranch = currentBranch

	return gitRepository, nil
}

func FindGitRoot(filePath string) (string, error) {
	// Walk up to the root directory in a portable way (works on Windows, macOS, Linux)
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	for {
		checkedPath := filepath.Join(abs, ".git")
		if _, err := os.Stat(checkedPath); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs { // reached filesystem root
			return "", fmt.Errorf("no git repository found")
		}
		abs = parent
	}
}

func getAbsolutePath(repoRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	repoRootAbsolute := strings.TrimSpace(string(out))
	return repoRootAbsolute, nil
}

func (git *GitRepository) ListAllCommitsSince(since string) ([]Commit, error) {
	cmd := exec.Command("git", "--no-pager", "log", "--pretty=format:# %h|%an|%ct|%ae", "--name-only", "--since="+since)
	cmd.Dir = git.Path

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var currentCommit Commit
	commits := make([]Commit, 0, 256)
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			commitInfos := strings.Split(line[2:], "|")
			if len(commitInfos) < 3 {
				log.Println("Invalid commit line in git log")
				continue
			}

			timestamp, err := strconv.Atoi(commitInfos[2])
			if err != nil {
				log.Println("Invalid timestamp in git log")
				continue
			}

			// Save previous commit if it has data
			if currentCommit.Hash != "" {
				commits = append(commits, currentCommit)
			}

			// The email is appended last, so a log without it still parses
			email := ""
			if len(commitInfos) > 3 {
				email = commitInfos[3]
			}

			currentCommit = Commit{
				Hash:      commitInfos[0],
				Author:    commitInfos[1],
				Email:     email,
				Timestamp: timestamp,
			}
			continue
		}

		if currentCommit.Hash == "" {
			continue
		}

		if line == "" {
			commits = append(commits, currentCommit)
			currentCommit = Commit{}
			continue
		}

		currentCommit.Files = append(currentCommit.Files, line)
	}

	// Don't forget the last commit if output doesn't end with empty line
	if currentCommit.Hash != "" {
		commits = append(commits, currentCommit)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return commits, nil
}

func (git *GitRepository) Checkout(commit string) error {

	if commit == "" {
		return fmt.Errorf("commit is empty")
	}

	// avoid to checkout the same commit
	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		return err
	}
	if currentBranch == commit {
		return nil
	}

	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = git.Path
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// ResolveRef resolves a branch, tag or commit expression to a full SHA.
func (git *GitRepository) ResolveRef(ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("ref is empty")
	}
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	cmd.Dir = git.Path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot resolve ref %q", ref)
	}
	return strings.TrimSpace(string(out)), nil
}

// MergeBase returns the best common ancestor between two commits.
func (git *GitRepository) MergeBase(a string, b string) (string, error) {
	cmd := exec.Command("git", "merge-base", a, b)
	cmd.Dir = git.Path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot compute merge-base between %q and %q", a, b)
	}
	return strings.TrimSpace(string(out)), nil
}

// AddWorktree checks out the given commit in a detached temporary worktree
// and returns its path. The caller is responsible for calling RemoveWorktree.
func (git *GitRepository) AddWorktree(commit string) (string, error) {
	dir, err := os.MkdirTemp("", "ast-metrics-worktree")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "worktree", "add", "--detach", "--force", dir, commit)
	cmd.Dir = git.Path
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("cannot create git worktree for %q: %s", commit, strings.TrimSpace(string(out)))
	}
	return dir, nil
}

// RemoveWorktree removes a worktree previously created with AddWorktree.
func (git *GitRepository) RemoveWorktree(dir string) error {
	if dir == "" {
		return nil
	}
	cmd := exec.Command("git", "worktree", "remove", "--force", dir)
	cmd.Dir = git.Path
	if err := cmd.Run(); err != nil {
		// best effort: prune metadata and remove the directory manually
		os.RemoveAll(dir)
		prune := exec.Command("git", "worktree", "prune")
		prune.Dir = git.Path
		prune.Run()
	}
	return nil
}

func (git *GitRepository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = git.Path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (git *GitRepository) RestoreFirstBranch() error {
	return git.Checkout(git.InitialBranch)
}
