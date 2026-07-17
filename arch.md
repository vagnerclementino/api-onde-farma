# Arquitetura de Projeto Go para AWS Lambda (REST API)

Este documento define os padrões arquiteturais, diretórios e melhores práticas para o desenvolvimento de APIs REST escritas em Go rodando sobre AWS Lambda (integradas ao API Gateway).

---

## 1. O que é Screaming Architecture?

O termo **Screaming Architecture** (Arquitetura Gritante), cunhado por Robert C. Martin (Uncle Bob), prega que a estrutura de pastas e a organização de código do seu projeto devem **gritar o seu domínio de negócio** (o propósito da aplicação), e não as ferramentas técnicas, frameworks ou banco de dados que ele utiliza.

Ao olhar a raiz de `/internal`, não devemos ver pastas como `controllers/`, `models/` ou `views/`. Em vez disso, devemos ver de imediato os domínios de negócio, como `pharmacy/`, `user/` ou `billing/`.

### 1.1 Screaming Architecture Pura (Sem Hexagonal) em Go
Muitas vezes, a tentativa de alinhar Screaming Architecture com Arquitetura Hexagonal (Ports and Adapters) introduz uma quantidade excessiva de subpastas (como `/domain`, `/application`, `/ports`, `/adapters/http`, `/adapters/repository`). 

Em Go, a melhor prática para projetos de pequeno a médio porte é **manter os pacotes planos (flat)**. Isso significa que tudo que diz respeito ao domínio de Farmácia reside diretamente dentro de `internal/pharmacy/`, sem subpastas. O próprio nome dos arquivos dentro do pacote descreve sua responsabilidade (`handlers.go`, `models.go`, `repository.go`).

---

## 2. Melhores Práticas para AWS Lambda em Go

Para obter a melhor performance, menor latência (cold starts reduzidos) e melhor aproveitamento financeiro na AWS, as seguintes práticas devem ser seguidas no desenvolvimento e compilação do código Go:

### 2.1 Otimização de Cold Starts (Inicialização Eficiente)
* **Warm-up Cache (Main):** Inicialize os recursos mais pesados — como leitura de variáveis de ambiente (`config.Load`), conexão com o banco de dados (`pgxpool.New`) e criação de clientes HTTP/AWS — na função `main()` do executável, **antes** de chamar `lambda.Start()`. A AWS reutiliza o ambiente de execução e as conexões globais criadas em `main` em chamadas quentes subsequentes (warm starts).
* **Compilação Otimizada:** Utilize flags específicas de compilação no `go build` para remover tabelas de símbolos e informações de depuração desnecessárias, reduzindo o tamanho do binário final (e consequentemente o tempo de carregamento da Lambda):
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o build/bootstrap ./cmd/lambda
  ```

### 2.2 Escolha da Arquitetura (Graviton / ARM64)
* Configure as funções Lambda na AWS para rodar em arquitetura **ARM64** (Graviton2/3). 
* Compilar o executável com `GOARCH=arm64` permite que a aplicação utilize a infraestrutura Graviton da AWS, que oferece até **20% melhor desempenho** com um custo **20% menor** comparado à arquitetura x86_64 tradicional.

### 2.3 Roteamento Standard & Desacoplamento via Proxy
* **Decoupling:** Evite ler ou validar estruturas de dados exclusivas da AWS Lambda (`events.APIGatewayV2HTTPRequest`) dentro dos seus pacotes de domínio. Seus handlers devem ser escritos usando o padrão nativo `http.HandlerFunc` (`w http.ResponseWriter, r *http.Request`).
* **AWS Lambda API Proxy:** Utilize a biblioteca oficial da AWS (`github.com/awslabs/aws-lambda-go-api-proxy`) apenas no ponto de entrada (`cmd/lambda/main.go`). Ela atua como um adaptador transparente, permitindo usar o roteador nativo `http.ServeMux` (Go 1.22+) ou frameworks como `chi`.
* **Portabilidade:** Ao usar `http.HandlerFunc`, você pode testar handlers usando o pacote nativo `net/http/httptest` ou iniciar um servidor convencional (`cmd/server/main.go`) rodando localmente de forma simples, sem necessidade de simuladores como o LocalStack.

### 2.4 Gerenciamento de Pools de Conexão (PostgreSQL)
* Devido à natureza efêmera e ao auto-escalonamento rápido da AWS Lambda (onde cada requisição simultânea pode subir uma nova instância isolada), o banco de dados pode facilmente sofrer esgotamento de conexões se o pool não for controlado.
* **Tamanho do Pool:** Defina o limite máximo de conexões por instância da Lambda para um valor bem baixo (ex: `MaxConns = 2` ou `MaxConns = 3` por pool).
* **RDS Proxy:** Para ambientes de alta concorrência em produção, é altamente recomendada a utilização do **AWS RDS Proxy** à frente do banco PostgreSQL, permitindo que a AWS gerencie o pool de forma eficiente e centralizada.

### 2.5 Logging Estruturado com `log/slog`
* Utilize o pacote nativo `log/slog` (introduzido no Go 1.21) para emitir logs formatados em **JSON** (`slog.NewJSONHandler`). 
* A AWS CloudWatch indexa nativamente formatos JSON, permitindo usar o CloudWatch Logs Insights para pesquisar registros de forma extremamente rápida através de chaves customizadas (ex: procurar logs por `request_id`, `cnpj` ou níveis específicos de erro).

---

## 3. Estrutura de Diretórios Proposta

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

## 4. Padrão de Código Base

### 4.1 Ponto de Entrada da Lambda (`cmd/lambda/main.go`)
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

### 4.2 Registrando e Processando Rotas (`internal/pharmacy/handlers.go`)
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

## 5. Vantagens Deste Modelo
1. **Rápido Desenvolvimento Local:** Você pode testar e debugar rotas localmente escrevendo um arquivo `cmd/server/main.go` que execute `http.ListenAndServe(":8080", mux)` sem necessidade de emuladores Lambda complexos.
2. **Alta Testabilidade:** Handlers HTTP padrões são facilmente testados via testes unitários usando o pacote padrão `net/http/httptest`.
3. **Organização Clara:** Não há acoplamento exagerado de camadas ou excesso de interfaces. O código é focado no domínio de negócio.
4. **Portabilidade:** Se futuramente a aplicação precisar rodar em containers (AWS ECS, Google Cloud Run, Kubernetes), nenhuma linha de lógica de negócio ou HTTP precisará ser alterada, bastando criar um novo ponto de entrada (`cmd/server/main.go`).
