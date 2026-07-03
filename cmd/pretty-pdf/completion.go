package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion script for the specified shell.
Run the output through your shell's completion mechanism to enable tab completion.

Bash:
  pretty-pdf completion bash > /etc/bash_completion.d/pretty-pdf

Zsh:
  pretty-pdf completion zsh > "${fpath[1]}/_pretty-pdf"

Fish:
  pretty-pdf completion fish > ~/.config/fish/completions/pretty-pdf.fish

PowerShell:
  pretty-pdf completion powershell > _pretty-pdf.ps1 & . .\_pretty-pdf.ps1
`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}
