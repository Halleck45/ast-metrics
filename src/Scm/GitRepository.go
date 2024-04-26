package Scm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepository struct {
	Path string
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

func (git *GitRepository) ListAllCommitsSince(since string) ([]string, error) {
	// Get all commits since one year (only sha1)
	cmd := exec.Command("git", "--no-pager", "log", "--pretty=format:%H", "--since="+since)
	cmd.Dir = git.Path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commits := strings.Split(string(out), "\n")
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
