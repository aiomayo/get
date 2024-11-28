package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AIO-Develope/get/internal/config"
	"github.com/AIO-Develope/get/internal/editor"
	"github.com/AIO-Develope/get/internal/repository"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var version = "dev"
var updateCheckURL = "https://api.github.com/repos/AIO-Develope/get/releases/latest"

func main() {
	rootCmd := &cobra.Command{
		Use:     "get",
		Short:   "Quickly open GitHub repositories",
		Version: version,
		Run:     runInteractive,
	}

	openCmd.Flags().BoolP("ignore-config", "i", false, "Ignore saved editor configuration and prompt for editor selection")
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(openCmd)
	if err := checkForUpdates(); err != nil {
		fmt.Println("Update check warning:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkForUpdates() error {
	if version == "dev" {
		return nil
	}

	resp, err := http.Get(updateCheckURL)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch latest release: status code %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	latestVersionStr := strings.TrimPrefix(release.TagName, "v")
	currentVersionStr := version

	latestVersion, err := semver.NewVersion(latestVersionStr)
	if err != nil {
		return fmt.Errorf("invalid latest version: %v", err)
	}

	currentVersion, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		return fmt.Errorf("invalid current version: %v", err)
	}

	if latestVersion.GreaterThan(currentVersion) {
		fmt.Printf("A new version is available: %s (current: %s)\n",
			latestVersionStr, currentVersion)
		fmt.Println("Run 'get upgrade' to update")
	}

	return nil
}

var openCmd = &cobra.Command{
	Use:   "get [repo-name]",
	Short: "Open a specific repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]

		ignoreConfig, _ := cmd.Flags().GetBool("ignore-config")

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		repos, err := repository.FindRepositories()
		if err != nil {
			fmt.Println("Error finding repositories:", err)
			os.Exit(1)
		}

		var selectedRepo *repository.Repository
		for i := range repos {
			if strings.Contains(strings.ToLower(repos[i].Name), strings.ToLower(repoName)) {
				selectedRepo = &repos[i]
				break
			}
		}

		if selectedRepo == nil {
			fmt.Println("Repository not found.")
			os.Exit(1)
		}

		var editorToUse string
		if !ignoreConfig {
			storedRepo, exists := cfg.Repositories[selectedRepo.Name]
			if exists && storedRepo.Editor != "" {
				editorToUse = storedRepo.Editor
			} else {
				editorToUse = chooseEditor(cfg, selectedRepo)
				cfg.UpdateRepositoryEditor(selectedRepo.Name, editorToUse)
				if err := cfg.Save(); err != nil {
					fmt.Println("Error saving config:", err)
				}
			}
		} else {
			editorToUse = chooseEditor(cfg, selectedRepo)
		}

		if err := selectedRepo.Open(editorToUse); err != nil {
			fmt.Println("Error opening repository:", err)
			os.Exit(1)
		}
	},
}
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the 'get' CLI",
	Run: func(cmd *cobra.Command, args []string) {
		executable, err := os.Executable()
		if err != nil {
			fmt.Println("Error getting the executable path:", err)
			os.Exit(1)
		}

		var confirm bool
		err = survey.AskOne(&survey.Confirm{
			Message: "Are you sure you want to uninstall the 'get' CLI?",
			Default: false,
		}, &confirm)
		if err != nil {
			fmt.Println("Error confirming uninstall:", err)
			os.Exit(1)
		}

		if !confirm {
			fmt.Println("Uninstall canceled.")
			return
		}

		switch runtime.GOOS {
		case "linux", "darwin", "windows":
			if err := os.Remove(executable); err != nil {
				fmt.Println("Error uninstalling:", err)
				os.Exit(1)
			}
			fmt.Println("Uninstalled successfully.")
		default:
			fmt.Println("Uninstall is not supported on this platform.")
			os.Exit(1)
		}
	},
}

func chooseEditor(cfg *config.Config, selectedRepo *repository.Repository) string {
	var editorToUse string
	installedEditors := editor.DetectInstalledEditors()
	installedEditors = append(installedEditors, "Other")

	err := survey.AskOne(&survey.Select{
		Message: "Select an editor:",
		Options: installedEditors,
		Default: cfg.DefaultEditor,
	}, &editorToUse)

	if err != nil {
		fmt.Println("Error selecting editor:", err)
		os.Exit(1)
	}

	if editorToUse == "Other" {
		err := survey.AskOne(&survey.Input{Message: "Enter the editor command:"}, &editorToUse)
		if err != nil {
			fmt.Println("Error entering editor:", err)
			os.Exit(1)
		}
	}

	var setDefaultEditor bool
	err = survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Do you want to set '%s' as the default editor for this repository?", editorToUse),
		Default: false,
	}, &setDefaultEditor)

	if err != nil {
		fmt.Println("Error asking about setting the editor:", err)
		os.Exit(1)
	}

	if setDefaultEditor {
		cfg.UpdateRepositoryEditor(selectedRepo.Name, editorToUse)
		if err := cfg.Save(); err != nil {
			fmt.Println("Error saving config:", err)
			os.Exit(1)
		}
	}

	return editorToUse
}

func runInteractive(cmd *cobra.Command, args []string) {
	repos, err := repository.FindRepositories()
	if err != nil {
		fmt.Println("Error finding repositories:", err)
		os.Exit(1)
	}

	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}

	var selectedRepoName string
	err = survey.AskOne(&survey.Select{
		Message: "Select a repository:",
		Options: repoNames,
	}, &selectedRepoName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	openCmd.Run(cmd, []string{selectedRepoName})
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to the latest version of the 'get' CLI",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var command *exec.Cmd

		switch runtime.GOOS {
		case "linux", "darwin":
			command = exec.Command("bash", "-c", "curl -fsSL https://get.aio-web.xyz/install.sh | bash")
		case "windows":
			command = exec.Command("powershell", "-Command", "iwr -useb https://get.aio-web.xyz/install.ps1 | iex")
		default:
			fmt.Printf("Upgrade not supported on %s\n", runtime.GOOS)
			os.Exit(1)
		}

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin

		if err = command.Run(); err != nil {
			fmt.Printf("Upgrade failed: %v\n", err)
			os.Exit(1)
		}
	},
}
