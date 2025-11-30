package command

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/pterm/pterm"
)

type DeployGithubOrganizationCommand struct {
	Organization string
	Token        string
	Branch       string
	WorkflowPath string
	IncludeForks bool
}

type GitHubRepo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	Fork          bool   `json:"fork"`
	Private       bool   `json:"private"`
}

type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

type GitHubRef struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Type string `json:"type"`
		Sha  string `json:"sha"`
		URL  string `json:"url"`
	} `json:"object"`
}

type GitHubPullRequest struct {
	Number  int    `json:"number"`
	State   string `json:"state"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

type DeployResult struct {
	RepoName string
	Status   string // "success", "skipped", "error"
	PRNumber int
	Error    string
}

func NewDeployGithubOrganizationCommand(organization, token, branch, workflowPath string, includeForks bool) *DeployGithubOrganizationCommand {
	if branch == "" {
		branch = "chore/ast-metrics-setup"
	}
	if workflowPath == "" {
		workflowPath = ".github/workflows/ast-metrics.yml"
	}
	return &DeployGithubOrganizationCommand{
		Organization: organization,
		Token:        token,
		Branch:       branch,
		WorkflowPath: workflowPath,
		IncludeForks: includeForks,
	}
}

func (c *DeployGithubOrganizationCommand) Execute() error {
	// Setup context with cancellation for Ctrl+C handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		pterm.Println()
		pterm.Warning.Println("Interrupted by user. Exiting...")
		cancel()
		os.Exit(0)
	}()

	// Welcome message with clear explanation
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("AST-Metrics Deployment")

	pterm.Println()

	// Explain what will happen
	pterm.DefaultBox.WithTitle("What's happening?").
		WithTitleTopCenter().
		WithRightPadding(2).
		WithLeftPadding(2).
		Println(
			"• We'll scan your organization for eligible repositories\n" +
				"• You choose which repos to deploy to\n" +
				"• A Pull Request will be opened on each selected repo\n" +
				"• You stay in control - merge when you're ready",
		)

	pterm.Println()

	// Token permissions info
	pterm.Info.Println("Required token permissions: repo (write), pull_requests (write), workflows (write)")
	pterm.Println()

	// Fetch repositories
	repos, err := c.fetchRepositories()
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Filter eligible repositories
	eligibleRepos := c.filterEligibleRepos(repos)

	if len(eligibleRepos) == 0 {
		pterm.Warning.Println("No eligible repositories found (excluding archived repos and those with existing workflows)")
		return nil
	}

	pterm.Success.Printf("Found %d eligible repositories\n", len(eligibleRepos))
	pterm.Println()

	// Convert to RepoItem for selection
	repoItems := make([]cli.RepoItem, len(eligibleRepos))
	for i, repo := range eligibleRepos {
		repoItems[i] = cli.RepoItem{
			Name:          repo.Name,
			FullName:      repo.FullName,
			DefaultBranch: repo.DefaultBranch,
			Selected:      true,
		}
	}

	// Ask user to select repositories
	selectedRepos := cli.AskUserToSelectRepos(repoItems, "", "Select repositories to deploy AST-Metrics:")

	if len(selectedRepos) == 0 {
		pterm.Warning.Println("No repositories selected")
		return nil
	}

	pterm.Println()

	// Convert back to GitHubRepo
	selectedGitHubRepos := make([]GitHubRepo, len(selectedRepos))
	for i, repoItem := range selectedRepos {
		for _, repo := range eligibleRepos {
			if repo.FullName == repoItem.FullName {
				selectedGitHubRepos[i] = repo
				break
			}
		}
	}

	// Process repositories
	results := []DeployResult{}
	spinner, _ := pterm.DefaultSpinner.Start("Opening Pull Requests...")

	for i, repo := range selectedGitHubRepos {
		// Check if context was cancelled (Ctrl+C)
		select {
		case <-ctx.Done():
			spinner.Stop()
			pterm.Warning.Println("Operation cancelled")
			return nil
		default:
		}

		spinner.UpdateText(fmt.Sprintf("%s (%d/%d)", repo.Name, i+1, len(selectedGitHubRepos)))
		result := c.processRepository(repo)
		results = append(results, result)
		time.Sleep(100 * time.Millisecond)
	}

	spinner.Stop()

	// Display summary
	c.displaySummary(results)

	return nil
}

func (c *DeployGithubOrganizationCommand) fetchRepositories() ([]GitHubRepo, error) {
	allRepos := []GitHubRepo{}
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=%d&page=%d&type=all", c.Organization, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+c.Token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
		}

		var repos []GitHubRepo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, err
		}

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		// Check if there are more pages
		if len(repos) < perPage {
			break
		}
		page++
	}

	return allRepos, nil
}

func (c *DeployGithubOrganizationCommand) filterEligibleRepos(repos []GitHubRepo) []GitHubRepo {
	eligible := []GitHubRepo{}

	for _, repo := range repos {
		// Skip archived
		if repo.Archived {
			continue
		}

		// Skip forks if not included
		if repo.Fork && !c.IncludeForks {
			continue
		}

		// Check if workflow already exists
		hasWorkflow, err := c.hasExistingWorkflow(repo)
		if err != nil {
			// Log error but continue
			continue
		}
		if hasWorkflow {
			continue
		}

		eligible = append(eligible, repo)
	}

	return eligible
}

func (c *DeployGithubOrganizationCommand) hasExistingWorkflow(repo GitHubRepo) (bool, error) {
	// Check if .github/workflows directory exists and contains ast-metrics workflow
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/.github/workflows", repo.FullName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil // Assume no workflow if we can't check
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return false, err
	}

	// Check if any file matches *ast-metrics*.yml pattern
	for _, content := range contents {
		if strings.Contains(strings.ToLower(content.Name), "ast-metrics") &&
			(strings.HasSuffix(content.Name, ".yml") || strings.HasSuffix(content.Name, ".yaml")) {
			return true, nil
		}
	}

	return false, nil
}

func (c *DeployGithubOrganizationCommand) processRepository(repo GitHubRepo) DeployResult {
	result := DeployResult{
		RepoName: repo.Name,
		Status:   "error",
	}

	// Step 1: Get default branch SHA
	defaultBranchSHA, err := c.getBranchSHA(repo, repo.DefaultBranch)
	if err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			result.Error = "missing permissions. Cannot get default branch SHA."
		} else {
			result.Error = fmt.Sprintf("failed to get branch SHA: %v", err)
		}
		return result
	}

	// Step 2: Check if branch already exists
	branchExists, err := c.branchExists(repo, c.Branch)
	if err != nil {
		result.Error = fmt.Sprintf("failed to check branch: %v", err)
		return result
	}

	if !branchExists {
		// Create new branch
		err = c.createBranch(repo, c.Branch, defaultBranchSHA)
		if err != nil {
			if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
				result.Error = "missing permissions. Cannot create branch."
			} else {
				result.Error = fmt.Sprintf("failed to create branch: %v", err)
			}
			return result
		}
	}

	// Step 3: Create workflow file
	err = c.createWorkflowFile(repo, c.Branch)
	if err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			result.Error = "missing permissions. Cannot create workflow file."
		} else {
			result.Error = fmt.Sprintf("failed to create workflow: %v", err)
		}
		return result
	}

	// Step 4: Check if PR already exists
	existingPR, err := c.findExistingPR(repo, c.Branch, repo.DefaultBranch)
	if err == nil && existingPR > 0 {
		result.Status = "skipped"
		result.PRNumber = existingPR
		result.Error = "PR already exists"
		return result
	}

	// Step 5: Create Pull Request
	prNumber, err := c.createPullRequest(repo, c.Branch, repo.DefaultBranch)
	if err != nil {
		// PR might already exist (race condition)
		if strings.Contains(err.Error(), "already exists") {
			result.Status = "skipped"
			result.Error = "PR already exists"
			return result
		}
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			result.Error = "missing permissions. Cannot create PR."
		} else {
			result.Error = fmt.Sprintf("failed to create PR: %v", err)
		}
		return result
	}

	result.Status = "success"
	result.PRNumber = prNumber
	return result
}

func (c *DeployGithubOrganizationCommand) getBranchSHA(repo GitHubRepo, branch string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/ref/heads/%s", repo.FullName, branch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get branch: %d - %s", resp.StatusCode, string(body))
	}

	var ref GitHubRef
	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return "", err
	}

	return ref.Object.Sha, nil
}

func (c *DeployGithubOrganizationCommand) branchExists(repo GitHubRepo, branchName string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/ref/heads/%s", repo.FullName, branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func (c *DeployGithubOrganizationCommand) createBranch(repo GitHubRepo, branchName, sha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs", repo.FullName)

	payload := map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": sha,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create branch: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *DeployGithubOrganizationCommand) createWorkflowFile(repo GitHubRepo, branch string) error {
	workflowContent := `name: AST Metrics
on:
  push:
  pull_request:

jobs:
  analyse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: halleck45/action-ast-metrics@v1
`

	// Encode content to base64
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo.FullName, c.WorkflowPath)

	// Check if file already exists to get its SHA (required for updates)
	existingSHA := ""
	checkUrl := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repo.FullName, c.WorkflowPath, branch)
	checkReq, err := http.NewRequest("GET", checkUrl, nil)
	if err == nil {
		checkReq.Header.Set("Authorization", "token "+c.Token)
		checkReq.Header.Set("Accept", "application/vnd.github.v3+json")
		checkClient := &http.Client{Timeout: 10 * time.Second}
		checkResp, err := checkClient.Do(checkReq)
		if err == nil {
			if checkResp.StatusCode == http.StatusOK {
				var existingContent GitHubContent
				if json.NewDecoder(checkResp.Body).Decode(&existingContent) == nil {
					existingSHA = existingContent.Sha
				}
			}
			checkResp.Body.Close()
		}
	}

	payload := map[string]interface{}{
		"message": "Add AST-Metrics workflow",
		"content": encodedContent,
		"branch":  branch,
	}

	// Include SHA if file exists (required for updates)
	if existingSHA != "" {
		payload["sha"] = existingSHA
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create file: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *DeployGithubOrganizationCommand) createPullRequest(repo GitHubRepo, headBranch, baseBranch string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls", repo.FullName)

	title := "Add AST-Metrics quality analysis workflow"
	body := `This PR adds the AST-Metrics workflow to automatically analyze
architecture, complexity and hotspots on each push.

No configuration needed.
Documentation: https://halleck45.github.io/ast-metrics/`

	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  headBranch,
		"base":  baseBranch,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)
		if strings.Contains(errorMsg, "already exists") {
			return 0, fmt.Errorf("PR already exists")
		}
		return 0, fmt.Errorf("failed to create PR: %d - %s", resp.StatusCode, errorMsg)
	}

	var pr GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return 0, err
	}

	return pr.Number, nil
}

func (c *DeployGithubOrganizationCommand) findExistingPR(repo GitHubRepo, headBranch, baseBranch string) (int, error) {
	// Search for existing PRs with the same head and base
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?head=%s:%s&base=%s&state=open",
		repo.FullName, repo.FullName, headBranch, baseBranch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil // Assume no PR if we can't check
	}

	var prs []GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return 0, err
	}

	if len(prs) > 0 {
		return prs[0].Number, nil
	}

	return 0, nil
}

func (c *DeployGithubOrganizationCommand) displaySummary(results []DeployResult) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgLightGreen)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("Results")
	pterm.Println()

	successCount := 0
	skippedCount := 0
	errorCount := 0

	for _, result := range results {
		switch result.Status {
		case "success":
			successCount++
			pterm.Success.Printf("✓ %s → PR #%d\n", result.RepoName, result.PRNumber)
		case "skipped":
			skippedCount++
			pterm.Warning.Printf("— %s (already exists)\n", result.RepoName)
		case "error":
			errorCount++
			pterm.Error.Printf("✗ %s: %s\n", result.RepoName, result.Error)
		}
	}

	pterm.Println()

	// Summary box
	summaryText := fmt.Sprintf(
		"✓ %d successful   — %d skipped   ✗ %d errors",
		successCount, skippedCount, errorCount,
	)

	if errorCount > 0 {
		pterm.Warning.Println(summaryText)
	} else {
		pterm.Success.Println(summaryText)
	}

	if successCount > 0 {
		pterm.Println()
		pterm.Info.Println("→ Review and merge the PRs when you're ready!")
	}
}
