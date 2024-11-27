package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"gitter/internal/config"
	"gitter/internal/editor"
	"gitter/internal/repository"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "gitter",
		Short:   "Quickly open GitHub repositories",
		Version: version,
		Run:     runInteractive,
	}

	rootCmd.PersistentFlags().Bool("ignore-config", false, "Do not use the preferred editor saved in the config")

	rootCmd.AddCommand(openCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var openCmd = &cobra.Command{
	Use:   "gitter [repo-name]",
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

func init() {
	openCmd.Flags().BoolP("ignore-config", "i", false, "Ignore saved editor configuration and prompt for editor selection")
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
