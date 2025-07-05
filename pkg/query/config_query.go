package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"reports-system/internal/domain/entities"
)

type ConfigQuery struct {
	config *entities.QueryConfig
	BaseQuery
}

func NewConfigQuery(config *entities.QueryConfig) entities.Query {
	return &ConfigQuery{
		config: config,
	}
}

func (q *ConfigQuery) Name() string {
	return q.config.Name
}

func (q *ConfigQuery) Description() string {
	return q.config.Description
}

func (q *ConfigQuery) Validate(params map[string]interface{}) error {
	for _, paramConfig := range q.config.Parameters {
		value, exists := params[paramConfig.Name]

		if paramConfig.Required && !exists {
			// Verificar se há valor padrão
			if paramConfig.Default != nil {
				params[paramConfig.Name] = q.resolveDefault(paramConfig.Default)
				continue
			}
			return fmt.Errorf("required parameter '%s' is missing", paramConfig.Name)
		}

		if !exists {
			if paramConfig.Default != nil {
				params[paramConfig.Name] = q.resolveDefault(paramConfig.Default)
			}
			continue
		}

		if err := q.validateParam(paramConfig, value); err != nil {
			return err
		}
	}

	return nil
}

func (q *ConfigQuery) validateParam(config entities.ParamConfig, value interface{}) error {
	switch config.Type {
	case "date":
		return q.validateDate(config, value)
	case "datetime":
		return q.validateDateTime(config, value)
	case "string":
		return q.validateString(config, value)
	case "int":
		return q.validateInt(config, value)
	case "float":
		return q.validateFloat(config, value)
	case "bool":
		return q.validateBool(config, value)
	case "enum":
		return q.validateEnum(config, value)
	default:
		return fmt.Errorf("unsupported parameter type: %s", config.Type)
	}
}

func (q *ConfigQuery) validateDate(config entities.ParamConfig, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("parameter '%s' must be a string date", config.Name)
	}

	format := "2006-01-02"
	if val, ok := config.Validation["format"]; ok {
		if formatStr, ok := val.(string); ok {
			format = q.convertDateFormat(formatStr)
		}
	}

	date, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("parameter '%s' must be in format %s", config.Name, format)
	}

	// Validar range
	if minVal, ok := config.Validation["min"]; ok {
		if minStr, ok := minVal.(string); ok {
			minDate, err := time.Parse(format, minStr)
			if err == nil && date.Before(minDate) {
				return fmt.Errorf("parameter '%s' must be after %s", config.Name, minStr)
			}
		}
	}

	return nil
}

func (q *ConfigQuery) validateString(config entities.ParamConfig, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("parameter '%s' must be a string", config.Name)
	}

	if regexVal, ok := config.Validation["regex"]; ok {
		if regexStr, ok := regexVal.(string); ok {
			regex, err := regexp.Compile(regexStr)
			if err != nil {
				return fmt.Errorf("invalid regex for parameter '%s'", config.Name)
			}
			if !regex.MatchString(str) {
				return fmt.Errorf("parameter '%s' does not match required pattern", config.Name)
			}
		}
	}

	return nil
}

func (q *ConfigQuery) validateInt(config entities.ParamConfig, value interface{}) error {
	var intVal int64

	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		intVal = int64(v)
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("parameter '%s' must be an integer", config.Name)
		}
		intVal = val
	default:
		return fmt.Errorf("parameter '%s' must be an integer", config.Name)
	}

	// Validar range
	if minVal, ok := config.Validation["min"]; ok {
		if min, ok := minVal.(float64); ok && intVal < int64(min) {
			return fmt.Errorf("parameter '%s' must be at least %v", config.Name, min)
		}
	}

	if maxVal, ok := config.Validation["max"]; ok {
		if max, ok := maxVal.(float64); ok && intVal > int64(max) {
			return fmt.Errorf("parameter '%s' must be at most %v", config.Name, max)
		}
	}

	return nil
}

func (q *ConfigQuery) validateFloat(config entities.ParamConfig, value interface{}) error {
	var floatVal float64

	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("parameter '%s' must be a number", config.Name)
		}
		floatVal = val
	default:
		return fmt.Errorf("parameter '%s' must be a number", config.Name)
	}

	// Validar range
	if minVal, ok := config.Validation["min"]; ok {
		if min, ok := minVal.(float64); ok && floatVal < min {
			return fmt.Errorf("parameter '%s' must be at least %v", config.Name, min)
		}
	}

	return nil
}

func (q *ConfigQuery) validateBool(config entities.ParamConfig, value interface{}) error {
	switch v := value.(type) {
	case bool:
		return nil
	case string:
		_, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("parameter '%s' must be a boolean", config.Name)
		}
	default:
		return fmt.Errorf("parameter '%s' must be a boolean", config.Name)
	}

	return nil
}

func (q *ConfigQuery) validateEnum(config entities.ParamConfig, value interface{}) error {
	enumValues, ok := config.Validation["values"]
	if !ok {
		return fmt.Errorf("enum parameter '%s' must have 'values' validation", config.Name)
	}

	values, ok := enumValues.([]interface{})
	if !ok {
		return fmt.Errorf("enum values must be an array")
	}

	valueStr := fmt.Sprintf("%v", value)
	for _, enumVal := range values {
		if fmt.Sprintf("%v", enumVal) == valueStr {
			return nil
		}
	}

	return fmt.Errorf("parameter '%s' must be one of: %v", config.Name, values)
}

func (q *ConfigQuery) validateDateTime(config entities.ParamConfig, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("parameter '%s' must be a string datetime", config.Name)
	}

	format := "2006-01-02 15:04:05"
	if val, ok := config.Validation["format"]; ok {
		if formatStr, ok := val.(string); ok {
			format = q.convertDateFormat(formatStr)
		}
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("parameter '%s' must be in format %s", config.Name, format)
	}

	return nil
}

func (q *ConfigQuery) convertDateFormat(format string) string {
	// Converter formatos comuns para Go
	replacements := map[string]string{
		"YYYY": "2006",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
	}

	result := format
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

func (q *ConfigQuery) resolveDefault(defaultValue interface{}) interface{} {
	if str, ok := defaultValue.(string); ok {
		switch str {
		case "now()":
			return time.Now().Format("2006-01-02")
		default:
			// Verificar se é uma expressão de data
			if strings.HasPrefix(str, "now(") && strings.HasSuffix(str, ")") {
				return q.evaluateDateExpression(str)
			}
		}
	}

	return defaultValue
}

func (q *ConfigQuery) evaluateDateExpression(expr string) string {
	// Exemplo: now(-30d) = 30 dias atrás
	inner := strings.TrimPrefix(expr, "now(")
	inner = strings.TrimSuffix(inner, ")")

	if inner == "" {
		return time.Now().Format("2006-01-02")
	}

	// Parse simple expressions like -30d, +7d, etc.
	re := regexp.MustCompile(`^([+-]?\d+)([dhmy])`)
	matches := re.FindStringSubmatch(inner)

	if len(matches) == 3 {
		amount, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		now := time.Now()
		switch unit {
		case "d":
			return now.AddDate(0, 0, amount).Format("2006-01-02")
		case "m":
			return now.AddDate(0, amount, 0).Format("2006-01-02")
		case "y":
			return now.AddDate(amount, 0, 0).Format("2006-01-02")
		}
	}

	return time.Now().Format("2006-01-02")
}

func (q *ConfigQuery) BuildQuery(params map[string]interface{}) (string, []interface{}) {
	query := q.config.Query
	var args []interface{}

	// Substituir parâmetros nomeados (@param) por placeholders numerados
	paramIndex := 1
	for paramName, paramValue := range params {
		placeholder := "@" + paramName
		if strings.Contains(query, placeholder) {
			query = strings.ReplaceAll(query, placeholder, fmt.Sprintf("$%d", paramIndex))
			args = append(args, paramValue)
			paramIndex++
		}
	}

	return query, args
}

func (q *ConfigQuery) TransformResults(columns []string, rows [][]interface{}) (interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, row := range rows {
		item := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				// Aplicar field mapping se definido
				fieldName := col
				if q.config.Output.FieldMapping != nil {
					if mappedName, exists := q.config.Output.FieldMapping[col]; exists {
						fieldName = mappedName
					}
				}

				item[fieldName] = row[i]
			}
		}
		result = append(result, item)
	}

	return result, nil
}

func (q *ConfigQuery) OutputFormats() []string {
	if len(q.config.Output.Formats) > 0 {
		return q.config.Output.Formats
	}
	return []string{"json"}
}

func (q *ConfigQuery) CacheTTL() time.Duration {
	if q.config.CacheTTL != "" {
		if duration, err := time.ParseDuration(q.config.CacheTTL); err == nil {
			return duration
		}
	}
	return 10 * time.Minute
}
