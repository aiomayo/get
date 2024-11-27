package editor

import (
	"os/exec"
)

var (
	CommonEditors = []string{
		"code",
		"subl",
		"atom",
		"vim",
		"emacs",
		"notepad++",
	}

	JetBrainsEditors = []string{
		"phpstorm",
		"goland",
		"idea",
		"webstorm",
		"pycharm",
		"clion",
		"datagrip",
		"rider",
		"rubymine",
		"android-studio",
	}
)

func DetectInstalledEditors() []string {
	var installedEditors []string

	for _, editor := range append(CommonEditors, JetBrainsEditors...) {
		if _, err := exec.LookPath(editor); err == nil {
			installedEditors = append(installedEditors, editor)
		}
	}

	return installedEditors
}

func IsValidEditor(editor string) bool {
	_, err := exec.LookPath(editor)
	return err == nil
}
