
# **Relatórios GoReport**

**Objetivo:**
Desenvolver uma API robusta para geração de relatórios com suporte a múltiplos formatos de saída, sistema de consultas seguras e extensível.

**1. Arquitetura Principal**
- Framework: Fiber v3
- Padrão de projeto: Clean Architecture (com separação clara de camadas)
- Sistema de injeção de dependências para componentes de banco de dados

**2. Sistema de Rotas Dinâmicas**
- Padrão: `/api/v1/reports/:report_id`
- Métodos: 
  - `GET` para consultas simples (com parâmetros via query string)
  - `POST` para consultas complexas (com parâmetros via JSON body)
- Registro automático de rotas baseado em queries cadastradas

**3. Camada de Banco de Dados**
- Configuração via environment variables:
  ```env
  DB_MAX_CONNS=10
  DB_IDLE_CONNS=5
  DB_CONN_TIMEOUT=30s
  ```
- Pool de conexões com health checks periódicos
- Suporte nativo para:
  ```go
  postgres.Open()
  mysql.Open()
  sqlite.Open()
  ```

**4. Sistema de Query - Detalhamento Técnico**
Interface completa:
```go
type Query interface {
    // Metadados
    Name() string
    Description() string
    
    // Validação
    Validate(params map[string]interface{}) error
    
    // Construção SQL
    BuildQuery(params map[string]interface{}) (query string, args []interface{})
    
    // Pós-processamento
    TransformResults(columns []string, rows [][]interface{}) (interface{}, error)
    
    // Configuração
    OutputFormats() []string // json, csv, xlsx
    CacheTTL() time.Duration
}
```

**5. Transformação de Resultados**
- Sistema de mapeamento dinâmico com:
  - Tipagem automática (detecção de tipos SQL -> Go)
  - Formatação condicional:
    ```go
    // Exemplo: formatação monetária
    if column == "valor" {
        return formatCurrency(value)
    }
    ```
  - Suporte a nested objects via JSONB (PostgreSQL)
  - Transformação de timezones

**6. Segurança Avançada**
- Validação de parâmetros com regras:
  ```go
  type ParamRule struct {
      Name     string
      Type     string // "date", "numeric", "string"
      Required bool
      Regex    string
      Min      interface{}
      Max      interface{}
  }
  ```
- Proteções:
  - Timeout padrão de 30s por consulta
  - Limite de 1000 linhas por padrão (configurável)
  - Query whitelisting para todas as consultas

**7. Sistema de Cache**
- Implementação multi-nível:
  ```go
  type CacheProvider interface {
      Get(key string) ([]byte, error)
      Set(key string, value []byte, ttl time.Duration) error
  }
  ```
- Implementações incluídas:
  - In-memory (cache local)
  - Redis (para distribuição)

**8. Gerenciamento de Conexões**
- Health check integrado:
  ```go
  type DBHealth struct {
      MaxConns    int
      OpenConns   int
      WaitCount   int64
      WaitTime    time.Duration
  }
  ```
- Reconexão automática
- Monitoramento via Prometheus metrics

**9. Exemplo Completo**
```go
// order_periodo.go
type OrderPeriodoQuery struct {
    BaseQuery
}

func (q *OrderPeriodoQuery) Name() string {
    return "order_periodo"
}

func (q *OrderPeriodoQuery) Validate(params map[string]interface{}) error {
    rules := []ParamRule{
        {"data_inicial", "date", true, "", nil, nil},
        {"data_final", "date", true, "", nil, nil},
        {"cliente_id", "string", false, "^[A-Z]{3}[0-9]{6}$", nil, nil},
    }
    return q.ValidateParams(params, rules)
}

func (q *OrderPeriodoQuery) BuildQuery(params map[string]interface{}) (string, []interface{}) {
    query := `SELECT 
                numero AS order_id,
                valor AS amount,
                cliente AS customer,
                data_order AS order_date
              FROM orders
              WHERE data_order BETWEEN $1 AND $2`
    
    if params["cliente_id"] != nil {
        query += " AND cliente_id = $3"
        return query, []interface{}{params["data_inicial"], params["data_final"], params["cliente_id"]}
    }
    
    return query, []interface{}{params["data_inicial"], params["data_final"]}
}

func (q *OrderPeriodoQuery) TransformResults(columns []string, rows [][]interface{}) (interface{}, error) {
    // Transformação personalizada aqui
    return q.DefaultTransform(columns, rows)
}
```

**10. Entregáveis Finais**
1. Código fonte organizado em:
   ```
   /internal
      /app
      /domain
      /infra
      /usecase
   /pkg
      /query
      /report
   ```

**Exemplo de Chamada:**
```bash
curl -X GET \
  "http://api/reports/order_periodo?data_inicial=2023-01-01&data_final=2023-12-31&format=json" \
  -H "Authorization: Bearer token123"
```

**Resposta Esperada:**
```json
{
  "metadata": {
    "report": "order_periodo",
    "params": {
      "data_inicial": "2023-01-01",
      "data_final": "2023-12-31"
    },
    "generated_at": "2023-11-20T10:00:00Z"
  },
  "data": [
    {
      "order_id": "ORD1001",
      "amount": 125.50,
      "customer": "John Doe",
      "order_date": "2023-05-15T00:00:00Z"
    }
  ]
}
```

**Diferenciais Implementados:**
1. Sistema de plugins para formatos adicionais (Excel, PDF)
2. Suporte a streaming para grandes conjuntos de dados
3. Auditoria automática de acesso
4. Versionamento de queries



## **Configuração Externa**

1. **Estrutura de Arquivos de Configuração:**
   - Cada relatório terá um arquivo JSON no formato:
     ```json
     {
       "name": "order_periodo",
       "description": "Relatório de pedidos por período",
       "query": "SELECT numero, valor, cliente FROM orders WHERE data_order BETWEEN $1 AND $2",
       "params": [
         {
           "name": "data_inicial",
           "type": "date",
           "required": true,
           "validation": {
             "format": "YYYY-MM-DD",
             "min": "2020-01-01"
           }
         },
         {
           "name": "data_final",
           "type": "date",
           "required": true,
           "default": "now()"
         }
       ],
       "output": {
         "formats": ["json", "csv"],
         "field_mapping": {
           "numero": "order_id",
           "valor": "amount"
         }
       }
     }
     ```

2. **Sistema de Carregamento:**
   - Diretório `config/reports/` para armazenar os JSONs
   - Monitoramento automático de alterações nos arquivos
   - Recarregamento dinâmico sem reiniciar o servidor

3. **Extensão da Interface Query:**
   ```go
   type QueryConfig struct {
       Name        string         `json:"name"`
       Description string         `json:"description"`
       Query       string        `json:"query"`
       Parameters  []ParamConfig `json:"params"`
       Output      OutputConfig  `json:"output"`
   }

   type ParamConfig struct {
       Name     string      `json:"name"`
       Type     string      `json:"type"` // date, string, int, float, bool
       Required bool        `json:"required"`
       Default interface{} `json:"default"`
       Validation map[string]interface{} `json:"validation"`
   }
   ```

4. **Exemplo de Implementação:**
   ```go
   func LoadQueriesFromConfig(path string) (map[string]Query, error) {
       // Implementação do carregador de queries
   }

   // Exemplo de uso:
   queries, err := LoadQueriesFromConfig("./config/reports/")
   ```

5. **Sistema de Transformação com Configuração:**
   - Suporte a templates nos arquivos JSON:
     ```json
     "query": "SELECT * FROM orders WHERE status = {{.status}}"
     ```
   - Pré-processamento de queries com variáveis

6. **Validação Avançada:**
   - Tipos suportados:
     - date (com formatos customizáveis)
     - datetime
     - string (com regex)
     - numeric (com faixas)
     - enum (lista de valores)

7. **Exemplo Completo de Uso:**

Estrutura de diretórios:
```
/config
   /reports
      order_periodo.json
      sales_report.json
      customer_orders.json
```

Arquivo `order_periodo.json`:
```json
{
  "name": "order_periodo",
  "version": "1.0",
  "description": "Relatório de pedidos por período",
  "query": "SELECT id, numero, valor, cliente, data_order FROM orders WHERE data_order BETWEEN @data_inicial AND @data_final",
  "params": [
    {
      "name": "data_inicial",
      "type": "date",
      "required": true,
      "description": "Data inicial no formato YYYY-MM-DD",
      "validation": {
        "format": "YYYY-MM-DD",
        "min": "2020-01-01"
      }
    },
    {
      "name": "data_final",
      "type": "date",
      "required": true,
      "default": "now()"
    },
    {
      "name": "cliente_id",
      "type": "string",
      "required": false,
      "validation": {
        "regex": "^[A-Z]{3}[0-9]{6}$"
      }
    }
  ],
  "output": {
    "formats": ["json", "csv"],
    "field_mapping": {
      "numero": "order_number",
      "valor": "amount",
      "cliente": "customer_name"
    },
    "metadata": {
      "author": "Finance Department",
      "refresh_rate": "daily"
    }
  }
}
```

8. **Funcionalidades Adicionais:**
   - Suporte a herança de configurações
   - Variáveis de ambiente nos arquivos JSON
   - Valores default calculados:
     ```json
     "default": "now(-30d)" // 30 dias atrás
     ```

9. **Sistema de Versionamento:**
   - Controle de versão dos arquivos de configuração
   - API para listar todas as queries disponíveis:
     ```
     GET /api/reports/metadata
     ```

10. **Segurança:**
    - Validação de sintaxe SQL nos arquivos
    - Restrição de tabelas/colunas acessíveis
    - Sistema de permissão por query

**Exemplo de Chamada API:**
```bash
curl -X POST \
  http://localhost:3000/api/reports/order_periodo \
  -H "Content-Type: application/json" \
  -d '{
    "params": {
      "data_inicial": "2023-01-01",
      "data_final": "2023-12-31"
    },
    "format": "csv"
  }'
```

**Vantagens da Abordagem:**
1. Separação clara entre código e configuração
2. Atualização de queries sem recompilar
3. Documentação embutida nos arquivos JSON
4. Fácil integração com sistemas de CI/CD
5. Versionamento independente para cada relatório
