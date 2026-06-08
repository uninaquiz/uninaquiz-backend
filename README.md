# uninaquiz-backend

API REST de geração e gerenciamento de quizzes educacionais com IA, desenvolvida em Go.

---

## Sumário

- [Visão Geral](#visão-geral)
- [Stack](#stack)
- [Arquitetura](#arquitetura)
- [Estrutura de Pastas](#estrutura-de-pastas)
- [Pré-requisitos](#pré-requisitos)
- [Configuração](#configuração)
- [Rodando o Projeto](#rodando-o-projeto)
  - [Com Docker Compose](#com-docker-compose)
  - [Localmente (sem Docker)](#localmente-sem-docker)
  - [Modo de desenvolvimento (hot reload)](#modo-de-desenvolvimento-hot-reload)
- [Migrações](#migrações)
- [Endpoints da API](#endpoints-da-api)
  - [Health](#health)
  - [Auth](#auth)
  - [Quiz](#quiz)
  - [Users](#users)
- [Testes](#testes)
  - [Testes Unitários](#testes-unitários)
  - [Testes de Integração](#testes-de-integração)
  - [Cobertura](#cobertura)
  - [Testes de Carga (k6)](#testes-de-carga-k6)
- [Geração de Mocks](#geração-de-mocks)
- [Build & Deploy](#build--deploy)
- [Linting](#linting)

---

## Visão Geral

O **uninaquiz-backend** é uma API que permite criar usuários, autenticar via JWT e gerar quizzes educacionais com questões de múltipla escolha usando a **API Gemini (Google AI)**. Os quizzes gerados podem ser salvos no histórico do usuário, consultados individualmente ou removidos.

---

## Stack

| Camada           | Tecnologia                                          |
|------------------|-----------------------------------------------------|
| Linguagem        | Go 1.26+                                            |
| Framework HTTP   | [Gin](https://github.com/gin-gonic/gin)             |
| Banco de Dados   | PostgreSQL 16                                       |
| ORM              | [GORM](https://gorm.io)                             |
| Migrações        | [Goose v3](https://github.com/pressly/goose)        |
| Autenticação     | JWT ([golang-jwt/jwt v5](https://github.com/golang-jwt/jwt)) + bcrypt |
| IA               | [Google Gemini API](https://ai.google.dev) via `google/generative-ai-go` |
| Testes           | `testing` + [testify](https://github.com/stretchr/testify) + [gomock](https://github.com/uber-go/mock) + [testcontainers-go](https://golang.testcontainers.org) |
| Testes de Carga  | [k6](https://k6.io)                                 |
| Config           | [envconfig](https://github.com/kelseyhightower/envconfig) |

---

## Arquitetura

O projeto segue **Clean Architecture** + **Hexagonal Architecture** + **DDD** + **CQRS**:

```
Infrastructure → Application → Domain
cmd            → tudo (apenas na composição de dependências)
```

- **Domain** — entidades puras, interfaces de repositório e erros de negócio (sentinel errors). Sem dependências externas.
- **Application** — casos de uso, commands/queries (CQRS), ports (input ports), DTOs, mappers e interfaces de serviços.
- **Infrastructure** — implementações concretas: repositórios GORM, modelos de banco, controllers Gin, middlewares, adapters (bcrypt, JWT, Gemini).
- **`cmd/`** — composition root: factories, wiring de dependências, inicialização da aplicação.

---

## Estrutura de Pastas

```
cmd/
  main.go
  config/
    environment.go          ← Carrega variáveis de ambiente
    factories/
      make_server.go        ← Composition root (DI manual)
internal/
  domain/
    entities/               ← Entidades puras (User, Quiz, QuizQuestion)
    repositories/           ← Interfaces IUserRepository, IQuizRepository
    errors/                 ← Sentinel errors de domínio
  application/
    commands/               ← DTOs de escrita (CreateUserCommand, GenerateQuizCommand…)
    queries/                ← DTOs de leitura (GetAllUsersQuery)
    dto/                    ← Structs de resposta (UserResponse, QuizHistoryResponse…)
    mappers/                ← Funções domain ↔ DTO
    ports/
      commands.go           ← Input ports de escrita
      queries.go            ← Input ports de leitura
    services/               ← Interfaces IHasher, ITokenService, IAIService
    adapters/               ← Implementações: bcrypt, JWT, Gemini
    usecases/               ← Lógica de aplicação (um arquivo por caso de uso)
  infrastructure/
    database/
      models/               ← Modelos GORM (UserModel, QuizModel, QuizQuestionModel)
      repositories/         ← Implementações GORM dos repositórios
    http/
      controllers/          ← Controllers Gin (AuthController, QuizController, UserController)
      middleware/           ← AuthMiddleware, CORSMiddleware
  mocks/                    ← Mocks gerados pelo mockgen
db/
  migrations/               ← Arquivos SQL do Goose
tests/
  integration/              ← Testes de integração com testcontainers
  load/                     ← Scripts k6 de carga
```

---

## Pré-requisitos

- Go 1.26+
- Docker e Docker Compose
- [k6](https://k6.io/docs/getting-started/installation/) (apenas para testes de carga)
- [Goose](https://github.com/pressly/goose) (instalado automaticamente via `make`)
- [mockgen](https://github.com/uber-go/mock) (instalado automaticamente via `make`)
- Uma chave de API válida do [Google Gemini](https://aistudio.google.com/app/apikey)

---

## Configuração

Crie um arquivo `.env` na raiz do projeto (ou exporte as variáveis de ambiente):

```env
DATABASE_URL=postgres://root:root@localhost:5432/db_uninaquiz?sslmode=disable
JWT_SECRET_KEY=your_super_secret_key
GEMINI_API_KEY=your_gemini_api_key
```

| Variável         | Padrão                                                                | Obrigatório |
|------------------|-----------------------------------------------------------------------|-------------|
| `DATABASE_URL`   | `postgres://root:root@localhost:5432/db_uninaquiz?sslmode=disable`   | ✅          |
| `JWT_SECRET_KEY` | `supersecret`                                                         | ❌          |
| `GEMINI_API_KEY` | —                                                                     | ✅          |

---

## Rodando o Projeto

### Com Docker Compose

Sobe a API, o PostgreSQL e o pgAdmin:

```bash
docker compose up -d
```

| Serviço   | URL                         |
|-----------|-----------------------------|
| API       | http://localhost:8080        |
| pgAdmin   | http://localhost:8081        |
| PostgreSQL| `localhost:5432`             |

Credenciais padrão do pgAdmin: `admin@example.com` / `admin`

### Localmente (sem Docker)

1. Suba apenas o banco:
```bash
docker compose up -d postgresql
```

2. Execute as migrações:
```bash
make migrate-up
```

3. Build e execução:
```bash
make api
```

### Modo de desenvolvimento (hot reload)

Utiliza o [air](https://github.com/air-verse/air) para recarregar automaticamente:

```bash
make api-dev
```

> O `make api-dev` instala o `air` automaticamente se não estiver presente.

---

## Migrations

```bash
# Criar uma nova migration
make create-migration create_nome_da_tabela

# Aplicar todas as migrations pendentes
make migrate-up

# Reverter a última migration
make migrate-down
```

As migrations ficam em `db/migrations/` e são gerenciadas pelo [Goose](https://github.com/pressly/goose). **Nunca edite migrations existentes** — crie novas para alterar o schema.

---

## Endpoints da API

### Health

| Método | Rota      | Auth | Descrição              |
|--------|-----------|------|------------------------|
| `GET`  | `/health` | ❌   | Verifica status da API |

### Auth

Base: `/api/auth`

| Método | Rota        | Auth | Descrição                                  |
|--------|-------------|------|--------------------------------------------|
| `POST` | `/register` | ❌   | Cria um novo usuário e retorna JWT         |
| `POST` | `/login`    | ❌   | Autentica e retorna JWT                    |
| `POST` | `/logout`   | ❌   | Encerra a sessão (stateless)               |

**`POST /api/auth/register`**
```json
// Request
{ "username": "john_doe", "password": "minhasenha123" }

// Response 201
{
  "token": "eyJ...",
  "user": { "id": "uuid", "username": "john_doe", "createdAt": "...", "updatedAt": "..." }
}
```

**`POST /api/auth/login`**
```json
// Request
{ "username": "john_doe", "password": "minhasenha123" }

// Response 200
{
  "token": "eyJ...",
  "user": { "id": "uuid", "username": "john_doe", "createdAt": "...", "updatedAt": "..." }
}
```

### Quiz

Base: `/api/quiz` — **todas as rotas exigem** `Authorization: Bearer <token>`

| Método   | Rota            | Descrição                                          |
|----------|-----------------|----------------------------------------------------|
| `POST`   | `/generate`     | Gera um quiz via Gemini AI                         |
| `GET`    | `/history`      | Lista o histórico de quizzes do usuário            |
| `GET`    | `/:id`          | Retorna um quiz específico com suas questões       |
| `POST`   | `/history`      | Salva o score de um quiz realizado                 |
| `DELETE` | `/history/:id`  | Remove um quiz do histórico                        |

**`POST /api/quiz/generate`**
```json
// Request
{ "topic": "photosynthesis", "difficulty": "easy" }
// difficulty: "easy" | "medium" | "hard"

// Response 200
{
  "id": "uuid",
  "topic": "photosynthesis",
  "difficulty": "easy",
  "questions": [
    {
      "id": "uuid",
      "position": 1,
      "text": "What is photosynthesis?",
      "options": ["Option A", "Option B", "Option C", "Option D"],
      "correctIndex": 0,
      "explanation": "..."
    }
  ]
}
```

**`POST /api/quiz/history`**
```json
// Request
{ "id": "quiz-uuid", "score": 8 }

// Response 200
{ "ok": true }
```

### Users

Base: `/api`

| Método | Rota     | Auth | Descrição                          |
|--------|----------|------|------------------------------------|
| `GET`  | `/users` | ❌   | Lista usuários com paginação       |

**`GET /api/users?page=1&limit=10`**
```json
// Response 200
{
  "users": [...],
  "total": 42,
  "page": 1,
  "limit": 10
}
```

---

## Testes

### Testes Unitários

Executam use cases e mappers com mocks (sem infraestrutura real):

```bash
# Rápido — sem race detector
make test

# Com race detector e saída verbosa
make test-unit
```

### Testes de Integração

Usam [testcontainers-go](https://golang.testcontainers.org) para subir um PostgreSQL real em Docker:

```bash
make test-integration
```

> **Requer Docker** em execução.

### Cobertura

Gera relatório HTML de cobertura:

```bash
make test-coverage
# Abre coverage.html no browser
```

### Testes de Carga (k6)

> **Requer [k6](https://k6.io/docs/getting-started/installation/) instalado** e a API rodando.

#### Auth (registro + login)

Simula até **100 VUs** fazendo register e login simultâneos:

```bash
k6 run tests/load/auth_load_test.js \
  -e BASE_URL=http://localhost:8080
```

| Stage     | Duração | VUs |
|-----------|---------|-----|
| Ramp-up   | 30s     | 10  |
| Sustain   | 1m      | 50  |
| Spike     | 30s     | 100 |
| Hold peak | 1m      | 100 |
| Ramp-down | 30s     | 0   |

**Thresholds:** `p(95) < 250ms`, `error_rate < 1%`, `p(99) login < 300ms`

#### Quiz (histórico + busca)

Simula usuários autenticados lendo histórico e consultando quizzes. A geração via IA é excluída por ser rate-limited pelo Gemini:

```bash
# 1. Cria um usuário de teste e captura o token
export JWT_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"loadtest_user","password":"LoadTest@123"}' | jq -r .token)

# 2. Executa o teste de carga
k6 run tests/load/quiz_load_test.js \
  -e BASE_URL=http://localhost:8080 \
  -e JWT_TOKEN=$JWT_TOKEN
```

| Stage     | Duração | VUs |
|-----------|---------|-----|
| Ramp-up   | 20s     | 5   |
| Sustain   | 1m      | 30  |
| Spike     | 30s     | 80  |
| Hold peak | 1m      | 80  |
| Ramp-down | 20s     | 0   |

**Thresholds:** `p(95) < 250ms`, `quiz_error_rate < 1%`, `p(95) history < 200ms`

---

## Geração de Mocks

Os mocks são gerados com [gomock](https://github.com/uber-go/mock) a partir das interfaces em `domain/repositories`, `application/services` e `application/ports`:

```bash
make generate-mocks
```

Os arquivos são gerados em `internal/mocks/`.

---

## Linting

```bash
make lint
```

Verifica formatação com `gofmt` e analisa possíveis bugs com `go vet`.

---
