package output

import "fmt"

func PrintBanner(version string) {
	banner := `
   ╔══════════════════════════════════════════════╗
   ║                                              ║
   ║   ██████╗  ██████╗       ██████╗ ██████╗ ███████╗
   ║  ██╔════╝ ██╔═══██╗      ██╔══██╗██╔══██╗██╔════╝
   ║  ██║  ███╗██║   ██║█████╗██████╔╝██║  ██║█████╗
   ║  ██║   ██║██║   ██║╚════╝██╔═══╝ ██║  ██║██╔══╝
   ║  ╚██████╔╝╚██████╔╝      ██║     ██████╔╝██║
   ║   ╚═════╝  ╚═════╝       ╚═╝     ╚═════╝ ╚═╝
   ║                                              ║
   ║         Beautiful PDFs from MDX              ║
   ╚══════════════════════════════════════════════╝`

	fmt.Println(PrimaryStyle.Render(banner))
	if version != "" && version != "dev" {
		fmt.Println("  " + MutedStyle.Render("v"+version))
	}
	fmt.Println()
}
