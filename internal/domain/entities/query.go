package entities

import (
	"time"
)

type Query interface {
	Name() string
	Description() string
	Validate(params map[string]interface{}) error
	BuildQuery(params map[string]interface{}) (query string, args []interface{})
	TransformResults(columns []string, rows [][]interface{}) (interface{}, error)
	OutputFormats() []string
	CacheTTL() time.Duration
}

type QueryConfig struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Query       string         `json:"query"`
	Parameters  []ParamConfig  `json:"params"`
	Output      OutputConfig   `json:"output"`
	Security    SecurityConfig `json:"security,omitempty"`
	CacheTTL    string         `json:"cache_ttl,omitempty"`
}

type ParamConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default,omitempty"`
	Description string                 `json:"description,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

type OutputConfig struct {
	Formats      []string               `json:"formats"`
	FieldMapping map[string]string      `json:"field_mapping,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type SecurityConfig struct {
	AllowedTables []string `json:"allowed_tables,omitempty"`
	MaxRows       int      `json:"max_rows,omitempty"`
	RequireAuth   bool     `json:"require_auth,omitempty"`
}

type ParamRule struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Regex    string      `json:"regex,omitempty"`
	Min      interface{} `json:"min,omitempty"`
	Max      interface{} `json:"max,omitempty"`
}

type ReportMetadata struct {
	Report      string                 `json:"report"`
	Params      map[string]interface{} `json:"params"`
	GeneratedAt time.Time              `json:"generated_at"`
	Format      string                 `json:"format"`
}

type ReportResponse struct {
	Metadata ReportMetadata `json:"metadata"`
	Data     interface{}    `json:"data"`
}
