package generator

import "strings"

const (
	srcModule        = "github.com/my-org/my-service"
	srcOrg           = "my-org"
	srcServiceName   = "my-service"
	srcServiceSnake  = "my_service"
	srcServicePascal = "MyService"
	srcDBName        = "my-service-db"
	srcBufOrg        = "my-org"
)

func applyReplacements(content string, cfg *ProjectConfig) string {
	type r struct{ old, new string }
	reps := []r{
		{srcModule, cfg.Module},
		{"buf.build/" + srcBufOrg + "/" + srcServiceName, "buf.build/" + cfg.BufOrg + "/" + cfg.ServiceName},
		{srcDBName, cfg.DBName},
		// snake with trailing underscore first to avoid partial matches
		{srcServiceSnake + "_", cfg.ServiceSnake + "_"},
		{srcServiceSnake, cfg.ServiceSnake},
		{srcServicePascal, cfg.ServicePascal},
		{srcServiceName, cfg.ServiceName},
		{srcOrg, cfg.OrgName},
		{"go 1.22.0", "go " + cfg.GoVersion + ".0"},
		{"My service", cfg.ServicePascal + " service"},
	}
	for _, rep := range reps {
		if rep.old != rep.new {
			content = strings.ReplaceAll(content, rep.old, rep.new)
		}
	}
	return content
}

func transformPath(path string, cfg *ProjectConfig) string {
	path = strings.TrimSuffix(path, ".tmpl")
	path = strings.ReplaceAll(path, srcServiceSnake+"_", cfg.ServiceSnake+"_")
	path = strings.ReplaceAll(path, srcServiceSnake, cfg.ServiceSnake)
	path = strings.ReplaceAll(path, srcServiceName, cfg.ServiceName)
	return path
}
