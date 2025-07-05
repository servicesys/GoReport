package query

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"reports-system/internal/domain/entities"
)

type ConfigLoader struct {
	configPath string
}

func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{configPath: configPath}
}

func (cl *ConfigLoader) LoadQueries() (map[string]entities.Query, map[string]entities.QueryConfig, error) {
	queries := make(map[string]entities.Query)
	configs := make(map[string]entities.QueryConfig)

	files, err := filepath.Glob(filepath.Join(cl.configPath, "*.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	for _, file := range files {
		config, err := cl.loadConfigFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config file %s: %w", file, err)
		}

		query := NewConfigQuery(config)
		queries[config.Name] = query
		configs[config.Name] = *config
	}

	return queries, configs, nil
}

func (cl *ConfigLoader) loadConfigFile(filename string) (*entities.QueryConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config entities.QueryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Validar configuração
	if err := cl.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func (cl *ConfigLoader) validateConfig(config *entities.QueryConfig) error {
	if config.Name == "" {
		return fmt.Errorf("query name is required")
	}

	if config.Query == "" {
		return fmt.Errorf("query is required")
	}

	// Validar SQL básico (prevenir injeção)
	if err := cl.validateSQL(config.Query); err != nil {
		return fmt.Errorf("invalid SQL: %w", err)
	}

	return nil
}

func (cl *ConfigLoader) validateSQL(query string) error {
	// Lista de palavras-chave perigosas
	dangerous := []string{
		"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "TRUNCATE",
		"GRANT", "REVOKE", "EXEC", "EXECUTE", "xp_", "sp_",
	}

	upperQuery := strings.ToUpper(query)
	for _, word := range dangerous {
		// Regex: \b => delimitador de palavra
		pattern := `\b` + regexp.QuoteMeta(strings.ToUpper(word)) + `\b`
		re := regexp.MustCompile(pattern)
		if re.MatchString(upperQuery) {
			return fmt.Errorf("dangerous SQL keyword detected: %s", word)
		}
	}

	return nil
}
