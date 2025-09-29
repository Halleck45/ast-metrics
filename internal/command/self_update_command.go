package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
)

type SelfUpdateCommand struct {
	releaseUrl     string
	currentVersion string
}

type GithubAsset struct {
	Name                 string
	Browser_download_url string
}
type GithubRelease struct {
	Assets []GithubAsset
	Name   string
}

func NewSelfUpdateCommand(currentVersion string) *SelfUpdateCommand {

	return &SelfUpdateCommand{
		releaseUrl:     "https://api.github.com/repos/Halleck45/ast-metrics/releases/latest",
		currentVersion: currentVersion,
	}
}

func (v *SelfUpdateCommand) Execute() error {

	latest, err := v.GetLatestRelease()
	if err != nil {
		return err
	}

	// get current architecture
	arch := runtime.GOARCH
	os := runtime.GOOS
	os = strings.Title(os)

	// for arch, amd64 is x86_64
	if arch == "amd64" {
		arch = "x86_64"
	}

	fmt.Println("Versions:")
	fmt.Printf("  Current: %s (your version)\n", v.currentVersion)
	fmt.Printf("  Latest: %s", latest.Name)
	fmt.Println()

	for _, asset := range latest.Assets {
		if asset.Name == fmt.Sprintf("ast-metrics_%s_%s", os, arch) {
			binaryUrl := asset.Browser_download_url

			fmt.Printf("Updating to %s (%s_%s) ...\n", latest.Name, os, arch)

			// Download the latest version of the binary
			resp, err := http.Get(binaryUrl)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Apply self update
			err = selfupdate.Apply(resp.Body, selfupdate.Options{})
			if err != nil {
				return err
			}

			fmt.Println("Update complete")
			return nil
		}
	}

	fmt.Printf("No update found for your platform (%s_%s)\n", os, arch)
	fmt.Println("Please download it manually from: https://github.com/Halleck45/ast-metrics/releases/latest")
	return nil
}

func (v *SelfUpdateCommand) GetLatestRelease() (*GithubRelease, error) {

	// get JSON from Github API
	resp, err := http.Get(v.releaseUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// unmarshal response to struct
	response := GithubRelease{}
	err = json.NewDecoder(resp.Body).Decode(&response)

	return &response, nil
}
