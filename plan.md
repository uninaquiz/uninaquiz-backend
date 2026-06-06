# Plano de Refatoração — uninaquiz-backend

> Análise baseada nos princípios de **SOLID**, **Clean Code/Architecture**, **DDD** e **CQRS**.
> Cada item contém: o problema, o arquivo afetado, o princípio violado e o passo a passo de correção.

---

## Sumário Executivo

| Prioridade | Qtd | Categoria |
|---|---|---|
| 🔴 Crítico | 5 | Bugs, violações de camada graves |
| 🟠 Maior | 7 | Erros de design, inconsistências de comportamento |
| 🟡 Menor | 9 | Nomeação, estilo, boas práticas |

---

## 🔴 CRÍTICO

---

### C-1 — Domain Entities com tags de infraestrutura (`json`)

**Arquivo(s):** `internal/domain/entities/user.go`, `internal/domain/entities/quiz.go`

**Princípio violado:** DDD (Domain Purity), Clean Architecture (Dependency Rule)

**Problema:**
Entidades de domínio contêm tags `json:"..."` — um detalhe de serialização HTTP que pertence à camada de infraestrutura/apresentação. O domínio não deve saber sobre JSON.

```go
// ❌ ATUAL — domain/entities/user.go
type User struct {
    ID        string
    Username  string
    Password  string
    CreatedAt time.Time `json:"created_at"` // ← infraestrutura no domínio
    UpdatedAt time.Time `json:"updated_at"` // ← infraestrutura no domínio
}

// ❌ ATUAL — domain/entities/quiz.go
type Quiz struct {
    ID         string          `json:"id"`        // ← infraestrutura no domínio
    UserID     string          `json:"user_id"`
    // ... todos os campos têm json tags
}
```

**Correção:**

**Passo 1:** Remover todas as tags `json` das entidades de domínio.

```go
// ✅ CORRETO — domain/entities/user.go
type User struct {
    ID        string
    Username  string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// ✅ CORRETO — domain/entities/quiz.go
type Quiz struct {
    ID         string
    UserID     string
    Topic      string
    Difficulty TQuizDifficulty
    Score      int
    Total      int
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Passo 2:** As tags `json` já existem corretamente nos DTOs de resposta em `internal/application/commands/` — nenhuma mudança necessária lá.

---

### C-2 — Repositórios usam entidades de domínio diretamente como modelos GORM (sem camada de modelo)

**Arquivo(s):** `internal/infrastructure/database/repositories/user_repository.go`, `internal/infrastructure/database/repositories/quiz_repository.go`

**Princípio violado:** DDD, DIP (Dependency Inversion), Clean Architecture

**Problema:**
Os repositórios de infraestrutura passam as entidades de domínio diretamente para o GORM (`r.db.Create(&quiz)`, `usr.db.Model(&domain.User{})`). Isso faz com que o GORM precise inferir nome de tabela, colunas e relacionamentos das entidades de domínio — acoplando o domínio à infraestrutura de persistência.

```go
// ❌ ATUAL — quiz_repository.go
func (r *QuizRepository) Create(ctx context.Context, quiz domain.Quiz) (*domain.Quiz, error) {
    if err := r.db.WithContext(ctx).Create(&quiz).Error; err != nil { // domínio passado ao ORM
        return nil, err
    }
    return &quiz, nil
}
```

**Correção:**

**Passo 1:** Criar a pasta `internal/infrastructure/database/models/` com modelos GORM separados.

```go
// ✅ NOVO — internal/infrastructure/database/models/user_model.go
package models

import (
    "time"
    domain "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type UserModel struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Username  string    `gorm:"column:username;uniqueIndex;not null"`
    Password  string    `gorm:"column:password;not null"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string { return "tb_users" }

func (m *UserModel) ToDomain() *domain.User {
    return &domain.User{
        ID:        m.ID,
        Username:  m.Username,
        Password:  m.Password,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
}

func UserToModel(u domain.User) *UserModel {
    return &UserModel{
        ID:        u.ID,
        Username:  u.Username,
        Password:  u.Password,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}
```

```go
// ✅ NOVO — internal/infrastructure/database/models/quiz_model.go
package models

import (
    "time"
    domain "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type QuizModel struct {
    ID         string    `gorm:"column:id;primaryKey"`
    UserID     string    `gorm:"column:user_id;not null;index"`
    Topic      string    `gorm:"column:topic;not null"`
    Difficulty string    `gorm:"column:difficulty;not null"`
    Score      int       `gorm:"column:score;not null"`
    Total      int       `gorm:"column:total;not null"`
    CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (QuizModel) TableName() string { return "tb_quizzes" }

func (m *QuizModel) ToDomain() *domain.Quiz {
    return &domain.Quiz{
        ID:         m.ID,
        UserID:     m.UserID,
        Topic:      m.Topic,
        Difficulty: domain.TQuizDifficulty(m.Difficulty),
        Score:      m.Score,
        Total:      m.Total,
        CreatedAt:  m.CreatedAt,
        UpdatedAt:  m.UpdatedAt,
    }
}

func QuizToModel(q domain.Quiz) *QuizModel {
    return &QuizModel{
        ID:         q.ID,
        UserID:     q.UserID,
        Topic:      q.Topic,
        Difficulty: string(q.Difficulty),
        Score:      q.Score,
        Total:      q.Total,
        CreatedAt:  q.CreatedAt,
        UpdatedAt:  q.UpdatedAt,
    }
}
```

**Passo 2:** Atualizar os repositórios para usar os modelos e mapear de/para domínio:

```go
// ✅ CORRETO — user_repository.go
func (r *UserRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
    model := models.UserToModel(user)
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        return nil, err
    }
    return model.ToDomain(), nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&models.UserModel{}).Where("username = ?", username).Count(&count).Error; err != nil {
        return false, err
    }
    return count > 0, nil
}
```

**Passo 3:** Remover o `db.Table("tb_users")` do construtor — a tabela agora é definida pelo `TableName()` do model.

```go
// ✅ CORRETO
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}
```

---

### C-3 — Bug lógico em `ToQuizDifficulty` (método sempre retorna erro)

**Arquivo(s):** `internal/domain/entities/quiz.go`

**Princípio violado:** Clean Code (Correctness), DDD (Domain Logic)

**Problema:**
`isValid()` valida os valores em inglês (`"easy"`, `"medium"`, `"hard"`), mas o `switch` em `ToQuizDifficulty` verifica os valores em português (`"fácil"`, `"médio"`, `"difícil"`). As strings nunca casam, então a função **sempre** retorna `errors.New("invalid difficulty")` — código morto e com bug.

```go
// ❌ ATUAL — BUGADO
func (qd TQuizDifficulty) ToQuizDifficulty(difficulty string) (TQuizDifficulty, error) {
    if !isValid(difficulty) { // valida "easy/medium/hard"
        return "", errors.New("invalid difficulty")
    }
    switch difficulty {
    case "fácil":   // ← nunca vai casar com "easy"
        return DifficultEasy, nil
    case "médio":   // ← nunca vai casar com "medium"
        return DifficultMedium, nil
    case "difícil": // ← nunca vai casar com "hard"
        return DifficultHard, nil
    }
    return "", errors.New("invalid difficulty") // sempre cai aqui
}
```

**Correção:**

**Passo 1:** Reescrever como função construtora pura (não como método na própria type):

```go
// ✅ CORRETO
func ParseDifficulty(difficulty string) (TQuizDifficulty, error) {
    switch TQuizDifficulty(difficulty) {
    case DifficultEasy, DifficultMedium, DifficultHard:
        return TQuizDifficulty(difficulty), nil
    }
    return "", fmt.Errorf("invalid difficulty: %s", difficulty)
}
```

**Passo 2:** Remover o método `ToQuizDifficulty` e a função `isValid` — substituídos pela função acima.

**Passo 3:** Verificar se `ToQuizDifficulty` é chamado em algum lugar (pelo `grep` — atualmente parece não ser usado).

---

### C-4 — `CreateUser` duplicado: `UserController` e `AuthController` compartilham o mesmo use case

**Arquivo(s):** `internal/infrastructure/http/controllers/user_controller.go`, `internal/infrastructure/http/controllers/auth_controller.go`, `cmd/config/factories/make_server.go`

**Princípio violado:** SRP (Single Responsibility), DRY

**Problema:**
`UserController.CreateUser` e `AuthController.Register` fazem exatamente a mesma coisa, usando o mesmo `createUserUseCase`. Isso expõe o endpoint `/api/users` (POST) e `/auth/register` (POST) para o mesmo fluxo — gerando duplicidade, inconsistência de tokens (6h vs 7h) e responsabilidade ambígua.

```go
// ❌ DUPLICADO — user_controller.go
response, err := uc.CreateUserUseCase.Run(c.Request.Context(), createUserCommand, 6*time.Hour) // 6h

// ❌ DUPLICADO — auth_controller.go
response, err := ctrl.createUserUseCase.Run(c.Request.Context(), cmd, 7*24*time.Hour) // 7*24h = DIFERENTE!
```

**Correção:**

**Passo 1:** Remover `CreateUserUseCase` de `UserController` e o endpoint `POST /api/users`.

```go
// ✅ CORRETO — user_controller.go
type UserController struct {
    getAllUsersUseCase ports.IGetAllUsersPort
}

func NewUserController(getAllUsersUseCase ports.IGetAllUsersPort, r *gin.RouterGroup) *UserController {
    usc := &UserController{getAllUsersUseCase: getAllUsersUseCase}
    usc.setupRoutes(r)
    return usc
}

func (uc *UserController) setupRoutes(r *gin.RouterGroup) {
    r.GET("/users", uc.GetUsers)
}
```

**Passo 2:** Manter o registro apenas em `AuthController` com um único valor de expiração.

**Passo 3:** Atualizar a factory em `make_server.go` para não passar `createUserUseCase` para `NewUserController`.

---

### C-5 — `NewGeminiAIAdapter` chama `log.Fatalf` internamente (responsabilidade do caller)

**Arquivo(s):** `internal/application/adapters/ai.go`

**Princípio violado:** SRP, Clean Code (Error Handling)

**Problema:**
O adaptador decide terminar o processo com `log.Fatalf` — uma decisão que pertence ao composition root (`cmd/`), não ao adaptador. Além disso, impede testes unitários do adaptador.

```go
// ❌ ATUAL
func NewGeminiAIAdapter(apiKey string) *GeminiAIAdapter {
    client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
    if err != nil {
        log.Fatalf("failed to create gemini client: %v", err) // ← decisão do caller
    }
    return &GeminiAIAdapter{client: client}
}
```

**Correção:**

**Passo 1:** Retornar `error` do construtor.

```go
// ✅ CORRETO
func NewGeminiAIAdapter(apiKey string) (*GeminiAIAdapter, error) {
    client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create gemini client: %w", err)
    }
    return &GeminiAIAdapter{client: client}, nil
}
```

**Passo 2:** Tratar o erro no `make_server.go`:

```go
// ✅ CORRETO — make_server.go
aiService, err := adapters.NewGeminiAIAdapter(cfg.GeminiAPIKey)
if err != nil {
    log.Fatalf("failed to initialize AI service: %v", err)
}
```

---

## 🟠 MAIOR

---

### M-1 — `commands` package mistura DTOs, Responses e funções de mapeamento

**Arquivo(s):** `internal/application/commands/user.go`, `internal/application/commands/quiz.go`

**Princípio violado:** SRP, CQRS

**Problema:**
O pacote `commands` deveria conter apenas structs de entrada (DTOs/Commands). Atualmente contém:
- Commands de entrada ✅
- Response structs (mistura com output) ⚠️
- Funções de mapeamento `ToUserEntity`, `ToUserResponse`, `ToQuizEntity`, `ToQuizHistoryResponse` ❌

As funções de mapeamento criam acoplamento entre comandos de entrada e entidades de domínio.

**Correção:**

**Passo 1:** Separar em subpacotes dentro de `application/`:

```
internal/application/
  commands/          → apenas structs de entrada (write side)
    user.go          → CreateUserCommand
    auth.go          → LoginCommand
    quiz.go          → SaveQuizHistoryCommand, GenerateQuizCommand
  queries/           → apenas structs de consulta (read side) — ver M-2
    user.go          → GetAllUsersQuery
    quiz.go          → GetQuizHistoryQuery
  dto/               → structs de resposta compartilhadas
    user.go          → UserResponse, CreateUserResponse, GetAllUsersResponse, LoginResponse
    quiz.go          → QuizHistoryResponse, GenerateQuizResponse
  mappers/           → funções de mapeamento (domínio ↔ DTO)
    user.go          → ToUserResponse, ToUserEntity
    quiz.go          → ToQuizHistoryResponse, ToQuizEntity
```

**Passo 2:** Atualizar todos os imports nos use cases e controllers.

---

### M-2 — Ausência de separação CQRS (Commands vs Queries)

**Arquivo(s):** `internal/application/ports/input.go`, `internal/application/commands/`

**Princípio violado:** CQRS

**Problema:**
Commands (escrita) e Queries (leitura) estão misturados no mesmo pacote e na mesma interface. CQRS exige segregação:
- **Commands** — mutam estado, retornam apenas confirmação/erro
- **Queries** — apenas leem, retornam dados, sem efeitos colaterais

```go
// ❌ MISTURADO — input.go
type ICreateUserPort interface { Run(...) (*commands.CreateUserResponse, error) } // command
type IGetAllUsersPort interface { Run(...) (*commands.GetAllUsersResponse, error) } // query
type ISaveQuizHistoryPort interface { Run(...) error }  // command
type IGetQuizHistoryPort interface { Run(...) } // query
```

**Correção:**

**Passo 1:** Separar os ports em dois arquivos:

```go
// ✅ CORRETO — internal/application/ports/commands.go
package ports

type ICreateUserPort interface {
    Run(ctx context.Context, input commands.CreateUserCommand, expirationTime time.Duration) (*dto.CreateUserResponse, error)
}

type ILoginUserPort interface {
    Run(ctx context.Context, input commands.LoginCommand, expirationTime time.Duration) (*dto.LoginResponse, error)
}

type ISaveQuizHistoryPort interface {
    Run(ctx context.Context, input commands.SaveQuizHistoryCommand, userID string) error
}

type IDeleteQuizHistoryPort interface {
    Run(ctx context.Context, id string, userID string) error
}

type IGenerateQuizPort interface {
    Run(ctx context.Context, input commands.GenerateQuizCommand, userID string) (*dto.GenerateQuizResponse, error)
}
```

```go
// ✅ CORRETO — internal/application/ports/queries.go
package ports

type IGetAllUsersPort interface {
    Run(ctx context.Context, input queries.GetAllUsersQuery) (*dto.GetAllUsersResponse, error)
}

type IGetQuizHistoryPort interface {
    Run(ctx context.Context, userID string) ([]dto.QuizHistoryResponse, error)
}
```

---

### M-3 — `QuizController` recebe `tokenService` que não usa diretamente (ISP violado)

**Arquivo(s):** `internal/infrastructure/http/controllers/quiz_controller.go`, `cmd/config/factories/make_server.go`

**Princípio violado:** ISP (Interface Segregation), SRP

**Problema:**
`QuizController` recebe `services.ITokenService` no construtor apenas para repassá-lo ao `middleware.AuthMiddleware`. O controller não usa o token service — cria acoplamento desnecessário.

```go
// ❌ ATUAL
func NewQuizController(
    ...,
    tokenService services.ITokenService, // ← controller não usa, só repassa
    r *gin.RouterGroup,
) *QuizController {
    ctrl := &QuizController{...}
    ctrl.setupRoutes(tokenService, r) // ← só para isso
    return ctrl
}
```

**Correção:**

**Passo 1:** Criar o middleware externamente na factory e injetá-lo pronto no `setupRoutes`:

```go
// ✅ CORRETO — quiz_controller.go
func NewQuizController(
    generateQuizUseCase    ports.IGenerateQuizPort,
    saveQuizHistoryUseCase ports.ISaveQuizHistoryPort,
    getQuizHistoryUseCase  ports.IGetQuizHistoryPort,
    deleteQuizHistoryUseCase ports.IDeleteQuizHistoryPort,
    authMiddleware gin.HandlerFunc, // ← já resolvido
    r *gin.RouterGroup,
) *QuizController {
    ctrl := &QuizController{...}
    ctrl.setupRoutes(authMiddleware, r)
    return ctrl
}

func (ctrl *QuizController) setupRoutes(auth gin.HandlerFunc, r *gin.RouterGroup) {
    r.POST("/generate", auth, ctrl.Generate)
    r.GET("/history", auth, ctrl.GetHistory)
    r.POST("/history", auth, ctrl.SaveHistory)
    r.DELETE("/history/:id", auth, ctrl.DeleteHistory)
}
```

**Passo 2:** Atualizar `make_server.go`:

```go
// ✅ CORRETO — make_server.go
authMiddleware := middleware.AuthMiddleware(tokenService)
controllers.NewQuizController(
    generateQuizUseCase,
    saveQuizHistoryUseCase,
    getQuizHistoryUseCase,
    deleteQuizHistoryUseCase,
    authMiddleware,
    quizGroup,
)
```

---

### M-4 — `DeleteHistory` retorna 404 para **todos** os erros (incluindo autorização negada)

**Arquivo(s):** `internal/infrastructure/http/controllers/quiz_controller.go`

**Princípio violado:** Clean Code (Error Handling), HTTP Semantics

**Problema:**
Quando o quiz existe mas o usuário não tem permissão para deletar, o use case retorna `"you are not allowed to delete this quiz history"`, mas o controller retorna 404 — semanticamente incorreto (deveria ser 403 Forbidden).

```go
// ❌ ATUAL
func (ctrl *QuizController) DeleteHistory(c *gin.Context) {
    if err := ctrl.deleteQuizHistoryUseCase.Run(c.Request.Context(), id, userID.(string)); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": err.Error()}) // ← 404 para tudo
        return
    }
}
```

**Correção:**

**Passo 1:** Adicionar erros tipados em `apperrors`:

```go
// ✅ CORRETO — apperrors/errors.go
var (
    ErrQuizNotFound     = errors.New("quiz history not found")
    ErrQuizForbidden    = errors.New("you are not allowed to delete this quiz history")
    ErrQuizAlreadyExists = errors.New("você já realizou um quiz sobre esse tema com essa dificuldade")
    ErrInvalidTopic      = errors.New("tema inválido: ...")
)
```

**Passo 2:** Usar `fmt.Errorf("%w", apperrors.ErrXxx)` nos use cases:

```go
// ✅ CORRETO — delete_quiz_history.go
if quiz == nil {
    return apperrors.ErrQuizNotFound
}
if quiz.UserID != userID {
    return apperrors.ErrQuizForbidden
}
```

**Passo 3:** Mapear corretamente no controller:

```go
// ✅ CORRETO — quiz_controller.go
func (ctrl *QuizController) DeleteHistory(c *gin.Context) {
    id := c.Param("id")
    userID, _ := c.Get("user_id")
    if err := ctrl.deleteQuizHistoryUseCase.Run(c.Request.Context(), id, userID.(string)); err != nil {
        switch {
        case errors.Is(err, apperrors.ErrQuizNotFound):
            c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
        case errors.Is(err, apperrors.ErrQuizForbidden):
            c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
        }
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
}
```

---

### M-5 — `apperrors` pertence à camada de domínio, não à aplicação

**Arquivo(s):** `internal/application/apperrors/errors.go`

**Princípio violado:** DDD, Clean Architecture (Dependency Rule)

**Problema:**
Erros de regra de negócio (`ErrQuizAlreadyExists`, `ErrInvalidTopic`) são erros do **domínio**, não da aplicação. Ao colocá-los em `application/apperrors`, qualquer camada que queira verificar esses erros precisa importar `application/` — o que pode levar a ciclos de dependência e viola o DDD.

**Correção:**

**Passo 1:** Criar `internal/domain/errors/errors.go`:

```go
// ✅ CORRETO — internal/domain/errors/errors.go
package domainerrors

import "errors"

var (
    ErrQuizAlreadyExists = errors.New("você já realizou um quiz sobre esse tema com essa dificuldade")
    ErrInvalidTopic      = errors.New("tema inválido: utilize um assunto educacional legítimo")
    ErrQuizNotFound      = errors.New("quiz history not found")
    ErrQuizForbidden     = errors.New("you are not allowed to delete this quiz history")
    ErrInvalidCredentials = errors.New("usuário ou senha incorretos")
)
```

**Passo 2:** Atualizar imports nos use cases e controllers para apontar para `domain/errors`.

**Passo 3:** Remover `internal/application/apperrors/`.

---

### M-6 — `QuizQuestion` definida dentro de `services/ai.go` (mistura de tipo e interface)

**Arquivo(s):** `internal/application/services/ai.go`, `internal/application/commands/quiz.go`

**Princípio violado:** SRP, ISP

**Problema:**
`QuizQuestion` é um DTO de domínio/resposta definido dentro do arquivo de interface de serviço. Isso faz com que `commands/quiz.go` importe `services` apenas para usar `QuizQuestion`:

```go
// ❌ ATUAL — commands/quiz.go
import "github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"

type GenerateQuizResponse struct {
    Questions []services.QuizQuestion `json:"questions"` // ← import de services em commands
}
```

**Correção:**

**Passo 1:** Mover `QuizQuestion` para `internal/application/dto/quiz.go` (após o M-1):

```go
// ✅ CORRETO — dto/quiz.go
type QuizQuestion struct {
    Text         string   `json:"text"`
    Options      []string `json:"options"`
    CorrectIndex int      `json:"correctIndex"`
    Explanation  string   `json:"explanation"`
}
```

**Passo 2:** Atualizar `IAIService` para usar o DTO:

```go
// ✅ CORRETO — services/ai.go
import "github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"

type IAIService interface {
    GenerateQuiz(ctx context.Context, topic string, difficulty string) ([]dto.QuizQuestion, error)
}
```

---

### M-7 — `SaveQuizHistoryUseCase` não valida se o quiz já existe antes de salvar

**Arquivo(s):** `internal/application/usecases/save_quiz_history.go`

**Princípio violado:** DDD (Business Rules), Clean Architecture

**Problema:**
`GenerateQuizUseCase` verifica se o quiz já existe antes de gerar (`ExistsByUserTopicAndDifficulty`), mas `SaveQuizHistoryUseCase` não faz a mesma verificação ao salvar. Um cliente malicioso pode chamar diretamente o endpoint `POST /quiz/history` com dados arbitrários, bypassando a verificação.

```go
// ❌ ATUAL — save_quiz_history.go
func (usc *SaveQuizHistoryUseCase) Run(ctx context.Context, input commands.SaveQuizHistoryCommand, userID string) error {
    quiz := commands.ToQuizEntity(input, userID)
    _, err := usc.quizRepository.Create(ctx, *quiz) // ← salva sem validar duplicata
    return err
}
```

**Correção:**

**Passo 1:** Adicionar verificação de duplicata antes do `Create`:

```go
// ✅ CORRETO — save_quiz_history.go
func (usc *SaveQuizHistoryUseCase) Run(ctx context.Context, input commands.SaveQuizHistoryCommand, userID string) error {
    exists, err := usc.quizRepository.ExistsByUserTopicAndDifficulty(ctx, userID, input.Topic, input.Difficulty)
    if err != nil {
        return err
    }
    if exists {
        return domainerrors.ErrQuizAlreadyExists
    }

    quiz := mappers.ToQuizEntity(input, userID)
    _, err = usc.quizRepository.Create(ctx, *quiz)
    return err
}
```

---

## 🟡 MENOR

---

### m-1 — Typo: `CreatedDAt` em `UserResponse`

**Arquivo(s):** `internal/application/commands/user.go`

**Princípio violado:** Clean Code (Naming)

**Correção:**
```go
// ❌
CreatedDAt time.Time `json:"createdAt"`

// ✅
CreatedAt time.Time `json:"createdAt"`
```
Atualizar também a função `ToUserResponse` que referencia `CreatedDAt`.

---

### m-2 — Nomes de constantes de dificuldade inconsistentes

**Arquivo(s):** `internal/domain/entities/quiz.go`

**Princípio violado:** Clean Code (Naming)

**Problema:** `DifficultEasy`, `DifficultMedium`, `DifficultHard` — o prefixo `Difficult` está errado (deveria ser `Difficulty`).

**Correção:**
```go
// ❌
const (
    DifficultEasy   TQuizDifficulty = "easy"
    DifficultMedium TQuizDifficulty = "medium"
    DifficultHard   TQuizDifficulty = "hard"
)

// ✅
const (
    DifficultyEasy   TQuizDifficulty = "easy"
    DifficultyMedium TQuizDifficulty = "medium"
    DifficultyHard   TQuizDifficulty = "hard"
)
```
Atualizar todos os usos dessas constantes.

---

### m-3 — Package name `domain` para o pacote de entidades (inconsistente com diretório `entities/`)

**Arquivo(s):** `internal/domain/entities/user.go`, `internal/domain/entities/quiz.go`

**Princípio violado:** Clean Code (Naming)

**Problema:** O diretório é `entities/` mas o `package` declarado é `domain`. Isso causa importações como `domain "github.com/.../entities"` — confuso e não convencional em Go.

**Correção:** Renomear o package para `entities`:
```go
// ❌
package domain

// ✅
package entities
```
Atualizar todos os imports que usam o alias `domain "...entities"` para importar diretamente sem alias ou com alias `entities`.

---

### m-4 — Códigos HTTP como literais inteiros no `UserController`

**Arquivo(s):** `internal/infrastructure/http/controllers/user_controller.go`

**Princípio violado:** Clean Code (Readability)

**Correção:** Substituir literais por constantes do pacote `net/http`:
```go
// ❌
c.JSON(400, gin.H{"message": err.Error()})
c.JSON(409, gin.H{"message": err.Error()})
c.JSON(201, response)
c.JSON(500, gin.H{"message": err.Error()})
c.JSON(200, response)

// ✅
c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
c.JSON(http.StatusCreated, response)
c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
c.JSON(http.StatusOK, response)
```
Adicionar import `"net/http"`.

---

### m-5 — Campos exportados em `UserController` vs campos privados em `AuthController`

**Arquivo(s):** `internal/infrastructure/http/controllers/user_controller.go`

**Princípio violado:** Clean Code (Consistency), Encapsulation

**Problema:**
```go
// ❌ — user_controller.go
type UserController struct {
    CreateUserUseCase  ports.ICreateUserPort  // Exportado
    GetAllUsersUseCase ports.IGetAllUsersPort // Exportado
}

// ✅ — auth_controller.go (correto)
type AuthController struct {
    createUserUseCase ports.ICreateUserPort  // privado
    loginUserUseCase  ports.ILoginUserPort   // privado
}
```

**Correção:** Tornar os campos de `UserController` privados:
```go
// ✅
type UserController struct {
    getAllUsersUseCase ports.IGetAllUsersPort
}
```

---

### m-6 — Mensagens de erro em idiomas mistos (PT + EN)

**Arquivo(s):** `internal/application/usecases/delete_quiz_history.go`, `internal/application/usecases/login_user.go`

**Princípio violado:** Clean Code (Consistency)

**Problema:**
- `"quiz history not found"` — inglês
- `"you are not allowed to delete this quiz history"` — inglês
- `"usuário ou senha incorretos"` — português
- Erros de apperrors em português
- Erros de middleware em inglês

**Correção:** Padronizar todos os erros em **inglês** (linguagem do código) ou todos em **português** (linguagem do produto). Recomendação: **inglês** para erros internos e de sistema; **português** apenas para mensagens voltadas ao usuário final (que podem ser controladas na camada de presentation/i18n).

Após M-5, as mensagens centralizadas em `domain/errors` facilitam a padronização.

---

### m-7 — `bcrypt.MaxCost` usado no hasher (extremamente lento em produção/testes)

**Arquivo(s):** `internal/application/adapters/hasher.go`

**Princípio violado:** Clean Code (Performance Awareness)

**Problema:** `bcrypt.MaxCost` (custo 31) torna o hash de senha proibitivamente lento — 30+ segundos por operação. O padrão de mercado é `bcrypt.DefaultCost` (custo 10-12).

```go
// ❌
bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MaxCost)

// ✅
bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

---

### m-8 — `GetAllUsersCommand` com `binding:"required"` mas controller silencia o erro

**Arquivo(s):** `internal/infrastructure/http/controllers/user_controller.go`, `internal/application/commands/user.go`

**Princípio violado:** Clean Code (Fail Fast), Consistency

**Problema:**
```go
// ❌ ATUAL — user_controller.go
if err := c.ShouldBindQuery(&getAllUsersCommand); err != nil {
    getAllUsersCommand.Page = 1   // silencia o erro e usa default
    getAllUsersCommand.Limit = 10 // nunca retorna 400 para query inválida
}
```

A validação de `binding:"required"` não faz sentido para query params com valores padrão. A binding e os defaults estão em conflito.

**Correção (duas opções):**

**Opção A — Remover `required` e usar apenas defaults:**
```go
// ✅ — commands/user.go
type GetAllUsersCommand struct {
    Page  int `form:"page,default=1"  binding:"min=1"`
    Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
```

**Opção B — Retornar erro 400 se binding falhar:**
```go
// ✅ — user_controller.go
if err := c.ShouldBindQuery(&getAllUsersCommand); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
    return
}
```

Recomendação: **Opção A** — mais amigável para APIs REST.

---

### m-9 — `GetQuizHistoryUseCase` sem paginação (risco de performance)

**Arquivo(s):** `internal/application/usecases/get_quiz_history.go`, `internal/domain/repositories/quiz.go`

**Princípio violado:** Clean Code (Scalability)

**Problema:** `FindAllByUserID` retorna todos os quizzes sem limite — pode ser problemático com usuários com muitos registros.

**Correção:**

**Passo 1:** Adicionar parâmetros de paginação à interface:
```go
// ✅ — domain/repositories/quiz.go
FindAllByUserID(ctx context.Context, userID string, page, limit int) ([]entities.Quiz, int64, error)
```

**Passo 2:** Criar `GetQuizHistoryQuery` com page/limit em `application/queries/`.

**Passo 3:** Atualizar use case, port, controller e repositório.

---

## Ordem de Execução Sugerida

Para minimizar quebras de compilação, siga esta sequência:

```
1. [C-3] Corrigir bug em ToQuizDifficulty
2. [m-7] Corrigir bcrypt.MaxCost → DefaultCost
3. [m-1] Corrigir typo CreatedDAt
4. [m-2] Renomear constantes DifficultX → DifficultyX
5. [M-5] Mover apperrors para domain/errors + adicionar novos erros (C-4 e M-4 dependem disso)
6. [C-1] Remover json tags das entidades de domínio
7. [m-3] Renomear package `domain` → `entities` nos arquivos de entidade + atualizar imports
8. [C-2] Criar database/models + atualizar repositórios
9. [C-4] Remover CreateUser de UserController + corrigir duplicata
10. [M-4] Corrigir HTTP codes em DeleteHistory usando errors tipados
11. [m-5] Tornar campos de UserController privados
12. [m-4] Substituir literais HTTP por constantes em UserController
13. [C-5] Retornar error de NewGeminiAIAdapter
14. [M-3] Mover tokenService para fora do QuizController
15. [M-1] Separar commands/queries/dto/mappers (maior refactor — fazer por último)
16. [M-2] Implementar CQRS ports em arquivos separados
17. [M-6] Centralizar QuizQuestion em dto/
18. [M-7] Adicionar validação de duplicata em SaveQuizHistoryUseCase
19. [m-6] Padronizar idioma das mensagens de erro
20. [m-8] Corrigir GetAllUsersCommand binding
21. [m-9] Adicionar paginação em GetQuizHistoryUseCase (opcional, por último)
```

---

## Checklist de Validação Pós-Refatoração

- [ ] `go build ./...` sem erros
- [ ] `go vet ./...` sem warnings
- [ ] Nenhuma importação de `infrastructure` dentro de `domain` ou `application`
- [ ] Nenhuma importação de `application` dentro de `domain`
- [ ] Todas as entidades de domínio sem tags de infraestrutura (`json`, `gorm`)
- [ ] Todos os modelos GORM com `TableName()` definido
- [ ] Todos os erros de negócio em `domain/errors`
- [ ] Commands e Queries em pacotes separados
- [ ] Controllers usando apenas constantes `http.StatusXxx`
- [ ] Nenhum `log.Fatalf` fora de `cmd/`
- [ ] `bcrypt.DefaultCost` em vez de `MaxCost`

