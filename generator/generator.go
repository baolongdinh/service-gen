package generator

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates
var templateFS embed.FS

const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func Run() error {
	printBanner()
	cfg, err := AskProjectConfig()
	if err != nil {
		return fmt.Errorf("collecting configuration: %w", err)
	}
	if err := confirmConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("\n%sGenerating project...%s\n", colorCyan, colorReset)
	if err := generate(cfg); err != nil {
		return fmt.Errorf("generating project: %w", err)
	}
	printSuccess(cfg)
	return nil
}

func generate(cfg *ProjectConfig) error {
	root := "templates/" + string(cfg.TemplateType)

	// Check if output directory already has files and warn
	if entries, err := os.ReadDir(cfg.OutputDir); err == nil && len(entries) > 0 {
		fmt.Printf("%s⚠️  Output directory %q already exists and is not empty.%s\n", colorYellow, cfg.OutputDir, colorReset)
		fmt.Print("Overwrite existing files? [y/N]: ")
		answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if a := strings.TrimSpace(strings.ToLower(answer)); a != "y" && a != "yes" {
			return fmt.Errorf("generation cancelled: output directory not empty")
		}
	}

	return fs.WalkDir(templateFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, root+"/")
		if rel == "" || rel == path {
			return nil
		}
		if d.Name() == "PLACEHOLDER" {
			return nil
		}
		outRel := transformPath(rel, cfg)
		out := filepath.Join(cfg.OutputDir, filepath.FromSlash(outRel))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, []byte(applyReplacements(string(data), cfg)), 0o644); err != nil {
			return err
		}
		fmt.Printf("  %s✓%s %s\n", colorGreen, colorReset, outRel)
		return nil
	})
}

func confirmConfig(cfg *ProjectConfig) error {
	label := map[TemplateType]string{
		TemplateTypeMicroservice: "microservice (Clean Architecture)",
		TemplateTypeMonolith:     "monolith-ddd-hexagonal (DDD + Hexagonal)",
	}[cfg.TemplateType]
	fmt.Println()
	fmt.Printf("%s── Project Configuration ──────────────────────────────%s\n", colorYellow, colorReset)
	fmt.Printf("  Template       : %s%s%s\n", colorCyan, label, colorReset)
	fmt.Printf("  Module         : %s%s%s\n", colorCyan, cfg.Module, colorReset)
	fmt.Printf("  Service name   : %s%s%s\n", colorCyan, cfg.ServiceName, colorReset)
	fmt.Printf("  Snake case     : %s%s%s\n", colorCyan, cfg.ServiceSnake, colorReset)
	fmt.Printf("  Pascal case    : %s%s%s\n", colorCyan, cfg.ServicePascal, colorReset)
	fmt.Printf("  Organization   : %s%s%s\n", colorCyan, cfg.OrgName, colorReset)
	fmt.Printf("  Buf org        : %s%s%s\n", colorCyan, cfg.BufOrg, colorReset)
	fmt.Printf("  Database name  : %s%s%s\n", colorCyan, cfg.DBName, colorReset)
	fmt.Printf("  Go version     : %s%s%s\n", colorCyan, cfg.GoVersion, colorReset)
	fmt.Printf("  Output dir     : %s%s%s\n", colorCyan, cfg.OutputDir, colorReset)
	fmt.Printf("%s───────────────────────────────────────────────────────%s\n", colorYellow, colorReset)
	fmt.Print("\nGenerate project with these settings? [Y/n]: ")
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if a := strings.TrimSpace(strings.ToLower(answer)); a == "n" || a == "no" {
		return fmt.Errorf("generation cancelled by user")
	}
	return nil
}

func printBanner() {
	fmt.Printf("%s\n  service-gen — Go service scaffolding tool\n%s\n\n", colorCyan, colorReset)
}

func printSuccess(cfg *ProjectConfig) {
	fmt.Printf("\n%s✅  Project generated at: %s%s\n", colorGreen, cfg.OutputDir, colorReset)
	fmt.Printf("%sNext steps:%s\n  cd %s\n  cp .env.example .env\n  make up\n  go test ./...\n  make gen\n",
		colorYellow, colorReset, cfg.OutputDir)
}
