package query

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"reports-system/internal/domain/entities"
)

type BaseQuery struct{}

func (b *BaseQuery) ValidateParams(params map[string]interface{}, rules []entities.ParamRule) error {
	for _, rule := range rules {
		value, exists := params[rule.Name]

		if rule.Required && !exists {
			return fmt.Errorf("required parameter '%s' is missing", rule.Name)
		}

		if !exists {
			continue
		}

		if err := b.validateParamType(rule, value); err != nil {
			return err
		}

		if rule.Regex != "" {
			if err := b.validateRegex(rule, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *BaseQuery) validateParamType(rule entities.ParamRule, value interface{}) error {
	switch rule.Type {
	case "date":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string date", rule.Name)
		}
		if _, err := time.Parse("2006-01-02", value.(string)); err != nil {
			return fmt.Errorf("parameter '%s' must be in format YYYY-MM-DD", rule.Name)
		}
	case "numeric":
		if _, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err != nil {
			return fmt.Errorf("parameter '%s' must be numeric", rule.Name)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string", rule.Name)
		}
	}

	return nil
}

func (b *BaseQuery) validateRegex(rule entities.ParamRule, value interface{}) error {
	regex, err := regexp.Compile(rule.Regex)
	if err != nil {
		return fmt.Errorf("invalid regex pattern for parameter '%s'", rule.Name)
	}

	if !regex.MatchString(fmt.Sprintf("%v", value)) {
		return fmt.Errorf("parameter '%s' does not match required pattern", rule.Name)
	}

	return nil
}

func (b *BaseQuery) DefaultTransform(columns []string, rows [][]interface{}) (interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, row := range rows {
		item := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				item[col] = row[i]
			}
		}
		result = append(result, item)
	}

	return result, nil
}
