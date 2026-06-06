# GitHub Copilot Instructions — uninaquiz-backend

## Stack & Contexto

- **Linguagem:** Go (1.26+)
- **Framework HTTP:** Gin (`github.com/gin-gonic/gin`)
- **Banco de dados:** PostgreSQL via GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- **Migrações:** Goose (`github.com/pressly/goose/v3`)
- **Auth:** JWT (`github.com/golang-jwt/jwt/v5`) + bcrypt (`golang.org/x/crypto/bcrypt`)
- **Módulo:** `github.com/EmanuelErnesto/uninaquiz-backend`

---

## Arquitetura: Clean Architecture + Hexagonal + DDD + CQRS

O projeto segue rigorosamente Clean Architecture, Hexagonal Architecture, DDD e CQRS. **Nunca misture responsabilidades entre camadas.**

### Mapa de camadas

```
cmd/                              → Composition Root (wiring, factories, main)
internal/
  domain/                         → Camada de Domínio (núcleo, sem dependências externas)
    entities/                     → Entidades de domínio (structs puras, lógica de negócio)
    repositories/                 → Interfaces de repositório (contratos, não implementações)
    errors/                       → Erros de negócio e de domínio (sentinel errors)
  application/                    → Camada de Aplicação (orquestração de casos de uso)
    usecases/                     → Casos de uso (regras de aplicação)
    commands/                     → DTOs de entrada — operações de escrita (CQRS write side)
    queries/                      → DTOs de entrada — operações de leitura (CQRS read side)
    dto/                          → Structs de resposta compartilhadas (output DTOs)
    mappers/                      → Funções de mapeamento domain ↔ DTO
    ports/
      commands.go                 → Input ports de escrita (ICreateUserPort, ISaveQuizHistoryPort…)
      queries.go                  → Input ports de leitura (IGetAllUsersPort, IGetQuizHistoryPort…)
    services/                     → Interfaces de serviços externos (IHasher, ITokenService, IAIService)
    adapters/                     → Implementações concretas dos serviços (bcrypt, jwt, gemini)
  infrastructure/                 → Camada de Infraestrutura (detalhes técnicos)
    database/
      models/                     → Modelos GORM (separados das entidades de domínio)
      repositories/               → Implementações de repositório (GORM)
    http/
      controllers/                → Controllers Gin (binding → use case → resposta)
      middleware/                 → Middlewares HTTP (auth, logging, etc.)
```

### Regra de dependência (estrita)

```
Infrastructure → Application → Domain
cmd            → tudo (só na composição)
```

- **Domain** não importa nada de `application/` ou `infrastructure/`.
- **Application** não importa nada de `infrastructure/` (apenas `application/services`, `domain/repositories` e `domain/errors`).
- **Infrastructure** implementa interfaces definidas em `domain/repositories` e `application/services`.
- **`cmd/`** é o único lugar onde o wiring acontece.

---

## Convenções de Código

### Domain (`internal/domain/`)

- Entidades são `struct` Go **puras** — **sem** tags `json:`, `gorm:`, `db:` ou qualquer outra tag de infraestrutura.
- Nomes de entidade no singular: `User`, `Quiz`.
- Interfaces de repositório prefixadas com `I`: `IUserRepository`, `IQuizRepository`.
- Erros de domínio ficam em `domain/errors/` como **sentinel errors** usando `errors.New`.
- Métodos de repositório semânticos ao domínio: `Create`, `FindByID`, `ExistsByUsername`, `Delete`.

```go
// ✅ Entidade correta — sem nenhuma tag de infraestrutura
type User struct {
    ID        string
    Username  string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// ✅ Interface de repositório correta
type IUserRepository interface {
    Create(ctx context.Context, user entities.User) (*entities.User, error)
    FindByUsername(ctx context.Context, username string) (*entities.User, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    GetAll(ctx context.Context, page, limit int) ([]entities.User, int64, error)
}

// ✅ Erros de domínio — domain/errors/errors.go
var (
    ErrUserAlreadyExists  = errors.New("user with this username already exists")
    ErrInvalidCredentials = errors.New("invalid username or password")
    ErrQuizNotFound       = errors.New("quiz history not found")
    ErrQuizForbidden      = errors.New("you are not allowed to perform this action")
    ErrQuizAlreadyExists  = errors.New("quiz with this topic and difficulty already exists for this user")
    ErrInvalidTopic       = errors.New("invalid topic: use a legitimate educational subject")
)
```

### Application (`internal/application/`)

#### CQRS — Commands vs Queries

Separe estritamente operações de **escrita** (commands) de **leitura** (queries):

```
commands/   → CreateUserCommand, LoginCommand, SaveQuizHistoryCommand, GenerateQuizCommand
queries/    → GetAllUsersQuery, GetQuizHistoryQuery
dto/        → UserResponse, CreateUserResponse, LoginResponse, QuizHistoryResponse, GenerateQuizResponse, QuizQuestion
mappers/    → ToUserResponse(), ToUserEntity(), ToQuizHistoryResponse(), ToQuizEntity()
```

```go
// ✅ Command (escrita) — commands/user.go
type CreateUserCommand struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=8,max=50"`
}

// ✅ Query (leitura) — queries/user.go
type GetAllUsersQuery struct {
    Page  int `form:"page,default=1"  binding:"min=1"`
    Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}

// ✅ DTO de resposta — dto/user.go
type UserResponse struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**Nunca** coloque funções de mapeamento dentro do pacote `commands/` ou `queries/` — use `mappers/`.

#### Ports separados por responsabilidade

```go
// ✅ ports/commands.go — input ports de escrita
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

// ✅ ports/queries.go — input ports de leitura
type IGetAllUsersPort interface {
    Run(ctx context.Context, input queries.GetAllUsersQuery) (*dto.GetAllUsersResponse, error)
}
type IGetQuizHistoryPort interface {
    Run(ctx context.Context, userID string) ([]dto.QuizHistoryResponse, error)
}
```

#### Use Cases

```go
// ✅ Use case correto
type CreateUserUseCase struct {
    userRepository repositories.IUserRepository
    hasher         services.IHasher
    tokenService   services.ITokenService
}

func NewCreateUserUseCase(
    repo repositories.IUserRepository,
    hasher services.IHasher,
    token services.ITokenService,
) *CreateUserUseCase {
    return &CreateUserUseCase{userRepository: repo, hasher: hasher, tokenService: token}
}

func (usc *CreateUserUseCase) Run(
    ctx context.Context,
    input commands.CreateUserCommand,
    expirationTime time.Duration,
) (*dto.CreateUserResponse, error) {
    // lógica de negócio aqui
}
```

### Infrastructure (`internal/infrastructure/`)

#### Modelos GORM (separados do domínio)

```go
// ✅ infrastructure/database/models/user_model.go
type UserModel struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Username  string    `gorm:"column:username;uniqueIndex;not null"`
    Password  string    `gorm:"column:password;not null"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string { return "tb_users" }

func (m *UserModel) ToDomain() *entities.User { ... }
func UserToModel(u entities.User) *UserModel { ... }
```

Repositórios **sempre** fazem o mapeamento `model → domain` antes de retornar. O nome da tabela é definido pelo método `TableName()` do model — **nunca** por `db.Table("...")` no construtor do repositório.

#### Controllers

Fluxo obrigatório: **bind input → chamar use case via port → formatar resposta**. Sem lógica de negócio.

```go
// ✅ Controller correto
func (ctrl *AuthController) Register(c *gin.Context) {
    var cmd commands.CreateUserCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    response, err := ctrl.createUserUseCase.Run(c.Request.Context(), cmd, 7*24*time.Hour)
    if err != nil {
        switch {
        case errors.Is(err, domainerrors.ErrUserAlreadyExists):
            c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
        default:
            log.Printf("register error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
        }
        return
    }
    c.JSON(http.StatusCreated, response)
}
```

- Sempre use constantes `net/http` (`http.StatusOK`, `http.StatusCreated`, etc.) — nunca literais inteiros.
- Campos de struct do controller devem ser **privados** (lowercase).
- Middlewares são **injetados prontos** no construtor — não criados internamente.

### Composition Root (`cmd/`)

- **Factories** em `cmd/config/factories/`: instanciam e injetam dependências manualmente.
- `log.Fatalf` é permitido **somente** em `cmd/` — nunca em adapters, use cases ou repositórios.
- Constructores que podem falhar devem retornar `(*T, error)`, deixando o `log.Fatalf` para a factory.
- Middlewares são criados nas factories e **injetados prontos** nos controllers:

```go
// ✅ Factory correta
authMiddleware := middleware.AuthMiddleware(tokenService)
controllers.NewQuizController(
    generateQuizUseCase,
    saveQuizHistoryUseCase,
    getQuizHistoryUseCase,
    deleteQuizHistoryUseCase,
    authMiddleware, // ← injetado pronto, não criado dentro do controller
    quizGroup,
)
```

---

## Erros e Respostas HTTP

- **Erros de domínio/negócio:** sentinel errors em `domain/errors/` via `errors.New("...")`.
- **Wrapping:** use `fmt.Errorf("contexto: %w", err)` para preservar a cadeia de erros.
- **Checagem:** use `errors.Is()` ou `errors.As()` — nunca compare strings de erro.
- **Nunca** use `panic` para erros de negócio esperados.
- Erros 500 nunca expõem detalhes de infraestrutura ao cliente HTTP — logue internamente, responda genericamente.
- Resposta de erro: `{ "message": "..." }`.

```go
// ❌ expõe detalhes de infra ao cliente
c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})

// ✅
log.Printf("internal error: %v", err)
c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
```

---

## Boas Práticas de Go

### Context

- `context.Context` é o **primeiro parâmetro** em toda operação de I/O (repositórios, serviços, adaptadores).
- Nunca armazene `context` em structs — passe sempre como parâmetro de função.
- Use `c.Request.Context()` nos controllers Gin para propagar o contexto HTTP.

```go
// ✅
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error)

// ❌ nunca armazene context em struct
type UserRepository struct { ctx context.Context }
```

### Erros

- Use `fmt.Errorf("%w", err)` para wrapping — habilita `errors.Is()` e `errors.As()` na cadeia.
- Nunca ignore erros em operações críticas (DB, hash, token) com `_`.
- Constructores que podem falhar retornam `(*T, error)`.

```go
// ✅
func NewGeminiAIAdapter(apiKey string) (*GeminiAIAdapter, error) {
    client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create gemini client: %w", err)
    }
    return &GeminiAIAdapter{client: client}, nil
}
```

### Interfaces (ISP)

- Interfaces devem ser **pequenas e focadas** — prefira 1-3 métodos.
- Defina interfaces **no pacote que as consome**, não no que as implementa.
- Nunca retorne um tipo concreto onde uma interface seria suficiente (exceto em `cmd/`).

### Segurança

- Nunca logue senhas, tokens ou API keys.
- Todos os secrets vêm de variáveis de ambiente via `cmd/config/environment.go`.

### Naming

- Packages: nomes curtos, lowercase, sem underscore. Evite "stutter" (`user.UserService` → ruim; `user.Service` → bom).
- Interfaces: prefixo `I` conforme convenção do projeto (`IUserRepository`, `IHasher`).
- Constantes de tipo: nomes semânticos e precisos (`DifficultyEasy`, não `DifficultEasy`).
- Evite abreviações obscuras em escopos amplos; em escopos de método curtos são aceitáveis.

### Testes

- Testes unitários em `_test.go` **na mesma pasta** do código.
- Use cases testados com **mocks** das interfaces — nunca com infraestrutura real.
- Controllers testados com `httptest.NewRecorder()` e `httptest.NewRequest()`.
- Prefira **table-driven tests** para múltiplos cenários.

```go
// ✅ table-driven test
func TestCreateUserUseCase_Run(t *testing.T) {
    tests := []struct {
        name    string
        input   commands.CreateUserCommand
        wantErr error
    }{
        {"success", commands.CreateUserCommand{Username: "john", Password: "pass1234"}, nil},
        {"duplicate", commands.CreateUserCommand{Username: "existing", Password: "pass1234"}, domainerrors.ErrUserAlreadyExists},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

### Padrões a Evitar

```go
// ❌ init() para lógica de negócio ou dependências
func init() { db = connectDB() }

// ❌ variáveis globais para estado ou dependências
var globalDB *gorm.DB

// ❌ goroutines sem mecanismo de cancelamento
go func() { doWork() }()

// ❌ defer em loops (acumula até o fim da função)
for _, item := range items {
    f, _ := os.Open(item)
    defer f.Close()
}

// ❌ panic para erros de negócio esperados
if user == nil { panic("user not found") }

// ❌ comparar string de erro
if err.Error() == "not found" { ... }

// ✅ usar sentinel errors
if errors.Is(err, domainerrors.ErrQuizNotFound) { ... }
```

---

## Migrações

- Usar: `make create-migration nome_semantico` (ex: `make create-migration create_users_table`).
- Nomenclatura semântica obrigatória: `create_users_table`, `add_email_column_to_users`, `drop_unused_index`.
- Nunca usar GORM `AutoMigrate` em produção.
- Migrações são **imutáveis** — criar novas para mudanças, nunca editar existentes.
- Migrações devem ser idempotentes: usar `IF NOT EXISTS`, `IF EXISTS`.

---

## Checklist Rápido ao Gerar Código

```
[ ] Entidades de domínio sem tags json/gorm/db
[ ] Erros de negócio em domain/errors/ como sentinel errors com errors.New()
[ ] Commands (escrita) e Queries (leitura) em pacotes separados
[ ] Ports de commands e queries em arquivos separados
[ ] Mappers em application/mappers/, não em commands/ ou queries/
[ ] DTOs de resposta em application/dto/
[ ] Modelos GORM em infrastructure/database/models/ com TableName()
[ ] Repositórios mapeiam model → domain antes de retornar
[ ] Repositórios não usam db.Table() no construtor
[ ] Controllers dependem de ports (interfaces), não de use cases concretos
[ ] Campos de struct do controller são privados (lowercase)
[ ] Controllers usam http.StatusXxx — nunca literais inteiros
[ ] Erros verificados com errors.Is() / errors.As()
[ ] Erros 500 não expõem detalhes de infraestrutura ao cliente
[ ] context.Context como primeiro parâmetro em todo I/O
[ ] log.Fatalf apenas em cmd/
[ ] Constructores que podem falhar retornam (*T, error)
[ ] Middlewares criados nas factories e injetados prontos nos controllers
[ ] Testes unitários com mocks, sem infraestrutura real
[ ] Table-driven tests para múltiplos cenários
```
