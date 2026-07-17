# Arquitetura de Projeto Go para AWS Lambda (REST API)

Este documento define os padrões arquiteturais, diretórios e boas práticas para o desenvolvimento de APIs REST escritas em Go rodando sobre AWS Lambda (integradas ao API Gateway).

---

## 1. Diretrizes de Arquitetura

### 1.1 Screaming Architecture (Coesão por Domínio)
A estrutura de diretórios deve deixar evidente quais são as regras de negócio e domínios da aplicação (ex: `pharmacy`, `user`, `order`), em vez de focar em termos de frameworks ou camadas puras.
* Cada domínio deve possuir seu próprio pacote flat dentro de `internal/` (ex: `internal/pharmacy/`).
* **Não utilize camadas aninhadas da arquitetura hexagonal** (como `domain/`, `application/`, `adapters/`). Coloque o código do domínio diretamente na raiz do pacote de domínio.

### 1.2 Roteamento Standard & Desacoplamento da AWS
Para evitar o acoplamento do código de negócio aos tipos da AWS (como `events.APIGatewayV2HTTPRequest`), o projeto deve utilizar a biblioteca padrão do Go (`net/http`) para o roteamento e tratamento de requisições:
* Os handlers HTTP devem usar a assinatura padrão `func(w http.ResponseWriter, r *http.Request)`.
* Utilize o roteador nativo do Go (`http.ServeMux`) com suporte a métodos e caminhos dinâmicos (disponível a partir do Go 1.22).
* Toda a tradução dos eventos do API Gateway para requisições `net/http` deve ocorrer apenas no ponto de entrada da Lambda (`cmd/lambda/main.go`), utilizando a biblioteca oficial `github.com/awslabs/aws-lambda-go-api-proxy/httpadapter`.

### 1.3 Inicialização de Recursos
* Lógica pesada de inicialização (leitura de variáveis de ambiente, criação de pools de banco de dados e conexões externas) deve ser executada na função `main()` do `cmd/lambda/main.go` (fora do loop do handler da Lambda) para otimizar os *warm starts* e reduzir o impacto de *cold starts*.

---

## 2. Estrutura de Diretórios Proposta

```text
api-onde-farma/
├── cmd/
│   ├── lambda/
│   │   └── main.go          # Ponto de entrada da Lambda AWS
│   └── server/
│       └── main.go          # Opcional: Servidor local HTTP convencional para desenvolvimento
├── db/                      # Arquivos de migração e schemas do banco de dados
├── internal/
│   ├── pharmacy/            # Exemplo de Domínio: Farmácias (Screaming Architecture)
│   │   ├── handlers.go      # Handlers HTTP Go padrão e registro de rotas
│   │   ├── models.go        # Structs de domínio, requests e responses (com tags JSON)
│   │   ├── repository.go    # Acesso a banco de dados (SQLC ou queries diretas)
│   │   └── service.go       # Lógica e regras de negócio da farmácia
│   ├── platform/            # Código técnico compartilhável (não-domínio)
│   │   ├── config/          # Carregamento de configurações / ENV
│   │   └── database/        # Inicialização e utilitários de banco
│   └── logger/              # Logger estruturado (slog)
```

---

## 3. Padrão de Código Base

### 3.1 Ponto de Entrada da Lambda (`cmd/lambda/main.go`)
Conecta o roteador padrão do Go (`http.ServeMux`) ao ambiente de execução da AWS Lambda usando o proxy:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/vagner/api-onde-farma/internal/pharmacy"
	"github.com/vagner/api-onde-farma/internal/platform/config"
	"github.com/vagner/api-onde-farma/internal/platform/database"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar config: %v", err)
	}

	dbPool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer dbPool.Close()

	// 1. Instancia o roteador padrão do Go (ServeMux)
	mux := http.NewServeMux()

	// 2. Registra as rotas injetando o banco
	pharmacy.RegisterRoutes(mux, dbPool)

	// 3. Inicializa o adaptador Proxy para o API Gateway V2 (HTTP API)
	adapter := httpadapter.NewV2(mux)

	// 4. Inicializa o handler do Lambda
	lambda.Start(func(lambdaCtx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return adapter.ProxyWithContext(lambdaCtx, req)
	})
}
```

### 3.2 Registrando e Processando Rotas (`internal/pharmacy/handlers.go`)
Escreva handlers HTTP comuns utilizando os padrões nativos do Go:

```go
package pharmacy

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	repo *Repository
}

func RegisterRoutes(mux *http.ServeMux, db DBConnection) {
	repo := NewRepository(db)
	h := &Handler{repo: repo}

	// Aproveita o roteamento moderno do Go 1.22+
	mux.HandleFunc("GET /v1/pharmacies", h.List)
	mux.HandleFunc("POST /v1/pharmacies", h.Create)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := r.URL.Query().Get("state")

	result, err := h.repo.ListByState(ctx, state)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro interno"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
```

---

## 4. Vantagens Deste Modelo
1. **Rápido Desenvolvimento Local:** Você pode testar e debugar rotas localmente escrevendo um arquivo `cmd/server/main.go` que execute `http.ListenAndServe(":8080", mux)` sem necessidade de emuladores Lambda complexos.
2. **Alta Testabilidade:** Handlers HTTP padrões são facilmente testados via testes unitários usando o pacote padrão `net/http/httptest`.
3. **Organização Clara:** Não há acoplamento exagerado de camadas ou excesso de interfaces. O código é focado no domínio de negócio.
4. **Portabilidade:** Se futuramente a aplicação precisar rodar em containers (AWS ECS, Google Cloud Run, Kubernetes), nenhuma linha de lógica de negócio ou HTTP precisará ser alterada, bastando criar um novo ponto de entrada (`cmd/server/main.go`).
