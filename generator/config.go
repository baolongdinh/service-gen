package generator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type TemplateType string

const (
	TemplateTypeMicroservice TemplateType = "microservice"
	TemplateTypeMonolith     TemplateType = "monolith-ddd-hexagonal"
)

type ProjectConfig struct {
	TemplateType  TemplateType
	Module        string
	ServiceName   string
	ServiceSnake  string
	ServicePascal string
	OrgName       string
	BufOrg        string
	DBName        string
	GoVersion     string
	OutputDir     string
}

func AskProjectConfig() (*ProjectConfig, error) {
	r := bufio.NewReader(os.Stdin)
	cfg := &ProjectConfig{}

	fmt.Printf("\n%sSelect template type:%s\n", colorCyan, colorReset)
	fmt.Println("  1) microservice    — Clean Architecture")
	fmt.Println("  2) monolith-ddd   — DDD + Hexagonal (Ports & Adapters)")
	fmt.Print("Choice [1]: ")
	if readLine(r) == "2" {
		cfg.TemplateType = TemplateTypeMonolith
	} else {
		cfg.TemplateType = TemplateTypeMicroservice
	}

	fmt.Print("Go module path (e.g. github.com/acme/user-service): ")
	cfg.Module = readLine(r)
	if cfg.Module == "" {
		return nil, fmt.Errorf("module path is required")
	}
	if err := validateModulePath(cfg.Module); err != nil {
		return nil, err
	}

	defaultName := lastSegment(cfg.Module)
	fmt.Printf("Service name (kebab-case) [%s]: ", defaultName)
	cfg.ServiceName = readLineDefault(r, defaultName)
	if err := validateServiceName(cfg.ServiceName); err != nil {
		return nil, err
	}

	defaultOrg := firstDomain(cfg.Module)
	fmt.Printf("Organization / GitHub org name [%s]: ", defaultOrg)
	cfg.OrgName = readLineDefault(r, defaultOrg)

	fmt.Printf("Buf.build org name [%s]: ", cfg.OrgName)
	cfg.BufOrg = readLineDefault(r, cfg.OrgName)

	fmt.Print("Go version [1.22]: ")
	cfg.GoVersion = readLineDefault(r, "1.22")

	defaultOut := "./" + cfg.ServiceName
	fmt.Printf("Output directory [%s]: ", defaultOut)
	cfg.OutputDir = readLineDefault(r, defaultOut)

	cfg.ServiceSnake = strings.ReplaceAll(cfg.ServiceName, "-", "_")
	cfg.ServicePascal = toPascal(cfg.ServiceName)
	cfg.DBName = cfg.ServiceName + "-db"
	return cfg, nil
}

func readLine(r *bufio.Reader) string {
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func readLineDefault(r *bufio.Reader, def string) string {
	if s := readLine(r); s != "" {
		return s
	}
	return def
}

func lastSegment(s string) string {
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

func firstDomain(module string) string {
	host := strings.SplitN(module, "/", 2)[0]
	parts := strings.SplitN(host, ".", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return host
}

func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name is required")
	}
	if !unicode.IsLower(rune(name[0])) {
		return fmt.Errorf("service name must start with a lowercase letter, got %q", name)
	}
	for _, c := range name {
		if !unicode.IsLower(c) && !unicode.IsDigit(c) && c != '-' {
			return fmt.Errorf("service name must be kebab-case (lowercase letters, digits, hyphens), got %q", name)
		}
	}
	return nil
}

func validateModulePath(module string) error {
	if module == "" {
		return fmt.Errorf("module path is required")
	}
	parts := strings.Split(module, "/")
	if len(parts) < 2 {
		return fmt.Errorf("module path must contain at least one slash (e.g. github.com/org/service), got %q", module)
	}
	if !strings.Contains(parts[0], ".") {
		return fmt.Errorf("module path host must contain a dot (e.g. github.com), got %q", parts[0])
	}
	return nil
}

func toPascal(s string) string {
	var b strings.Builder
	for _, word := range strings.Split(s, "-") {
		if len(word) > 0 {
			b.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
	}
	return b.String()
}
