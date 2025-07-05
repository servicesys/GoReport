package usecase

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"reports-system/internal/domain/entities"
	"reports-system/pkg/query"
)

type ReportService struct {
	db          entities.Database
	cache       entities.CacheProvider
	queries     map[string]entities.Query
	queriesConf map[string]entities.QueryConfig
	loader      *query.ConfigLoader
}

func NewReportService(db entities.Database, cache entities.CacheProvider, configPath string) *ReportService {
	service := &ReportService{
		db:          db,
		cache:       cache,
		queries:     make(map[string]entities.Query),
		queriesConf: make(map[string]entities.QueryConfig),
		loader:      query.NewConfigLoader(configPath),
	}

	// Carregar queries do diretório de configuração
	error := service.LoadQueries()
	if error != nil {
		fmt.Printf("Error loading queries: %v\n", error)
		panic(error)
	}

	return service
}

func (s *ReportService) LoadQueries() error {
	queries, queriesConf, err := s.loader.LoadQueries()
	if err != nil {
		return fmt.Errorf("failed to load queries: %w", err)
	}

	s.queries = queries
	s.queriesConf = queriesConf
	return nil
}

func (s *ReportService) RegisterQuery(q entities.Query) {
	s.queries[q.Name()] = q
}

func (s *ReportService) GetReport(reportID string, params map[string]interface{}, format string) (*entities.ReportResponse, error) {
	query, exists := s.queries[reportID]
	if !exists {
		return nil, fmt.Errorf("report '%s' not found", reportID)
	}

	if err := query.Validate(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Verificar cache
	cacheKey := s.generateCacheKey(reportID, params)
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var response entities.ReportResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			return &response, nil
		}
	}

	// Executar query
	sqlQuery, args := query.BuildQuery(params)
	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var allRows [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		allRows = append(allRows, values)
	}

	// Transformar dados
	data, err := query.TransformResults(columns, allRows)
	if err != nil {
		return nil, fmt.Errorf("transformation error: %w", err)
	}

	response := &entities.ReportResponse{
		Metadata: entities.ReportMetadata{
			Report:      reportID,
			Params:      params,
			GeneratedAt: time.Now(),
			Format:      format,
		},
		Data: data,
	}

	// Salvar no cache
	if responseBytes, err := json.Marshal(response); err == nil {
		s.cache.Set(cacheKey, responseBytes, query.CacheTTL())
	}

	return response, nil
}

func (s *ReportService) generateCacheKey(reportID string, params map[string]interface{}) string {
	paramBytes, _ := json.Marshal(params)
	hash := md5.Sum(append([]byte(reportID), paramBytes...))
	return fmt.Sprintf("report:%x", hash)
}

func (s *ReportService) GetAvailableReports() map[string]interface{} {
	reports := make(map[string]interface{})
	for name, query := range s.queries {
		reports[name] = map[string]interface{}{
			"name":        query.Name(),
			"description": query.Description(),
			"formats":     query.OutputFormats(),
			"query":       s.queriesConf[name].Query,
			"params":      s.queriesConf[name].Parameters,
		}
		//reports[query.] = query.Description()
	}
	return reports
}

func (s *ReportService) GetReportMetadata() map[string]interface{} {
	metadata := make(map[string]interface{})

	/*
		for name, query := range s.queries {
			if configQuery, ok := query.(*query.ConfigQuery); ok {
				metadata[name] = map[string]interface{}{
					"name":        configQuery.Name(),
					"description": configQuery.Description(),
					"formats":     configQuery.OutputFormats(),
					"cache_ttl":   configQuery.CacheTTL().String(),
				}
			} else {
				metadata[name] = map[string]interface{}{
					"name":        query.Name(),
					"description": query.Description(),
					"formats":     query.OutputFormats(),
					"cache_ttl":   query.CacheTTL().String(),
				}
			}
		}*/

	return metadata
}

func (s *ReportService) ReloadQueries() error {
	return s.LoadQueries()
}
