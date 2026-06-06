# Code Review Instructions — uninaquiz-backend

Este documento define os critérios de revisão de código para o projeto `uninaquiz-backend`.
Toda PR deve ser avaliada com base nestes pontos antes de ser aprovada.

---

## 1. Respeito às Camadas (Clean Arch / Hexagonal / DDD)

### ❌ Rejeitar se:
- Código de `domain/` importar qualquer coisa de `application/` ou `infrastructure/`.
- Código de `application/` importar qualquer coisa de `infrastructure/` (ex: GORM, Gin, pgx).
- Um controller executar lógica de negócio diretamente (ex: hash de senha, validação de regra de negócio).
- Um use case importar `gin`, `net/http` ou qualquer detalhe HTTP.
- Injeção de dependência acontecendo fora de `cmd/config/factories/`.
- `log.Fatalf` ou `os.Exit` fora de `cmd/`.
- Middleware sendo criado dentro de um controller ou use case (deve ser criado nas factories).

### ✅ Verificar:
- A regra de dependência flui em uma única direção: `Infrastructure → Application → Domain`.
- Cada arquivo está no pacote correto conforme a camada.
- Factories em `cmd/config/factories/` são o único ponto de composição.
- Middlewares são criados nas factories e injetados prontos nos construtores dos controllers.

---

## 2. Entidades de Domínio (DDD — Domain Purity)

### ❌ Rejeitar se:
- Entidades de `domain/entities/` possuírem qualquer tag de infraestrutura (`json:`, `gorm:`, `db:`, `form:`).
- Lógica de validação de formato estiver no controller ao invés do domínio.
- Entidades de domínio forem usadas diretamente como request/response HTTP ou como modelos GORM.
- Erros de negócio estiverem definidos fora de `domain/errors/` (ex: em `application/apperrors/`).
- Métodos de domínio possuírem bugs lógicos (ex: validar valor em inglês mas comparar em português).

### ✅ Verificar:
- Entidades são structs Go puras, sem dependências externas.
- Regras de negócio intrínsecas à entidade estão em métodos da própria entidade.
- Erros de negócio são **sentinel errors** em `domain/errors/` com `errors.New(...)`.
- Constantes de tipo têm nomes precisos e semânticos (ex: `DifficultyEasy`, não `DifficultEasy`).
- Package name em `domain/entities/` é `entities`, não `domain`.

---

## 3. CQRS — Separação de Commands e Queries

### ❌ Rejeitar se:
- Commands (escrita) e Queries (leitura) estiverem no mesmo pacote (`commands/`).
- Funções de mapeamento (`ToXxxEntity`, `ToXxxResponse`) estiverem dentro de `commands/` ou `queries/`.
- DTOs de resposta estiverem misturados com DTOs de entrada no mesmo arquivo/pacote.
- Input ports de escrita e leitura estiverem no mesmo arquivo (`ports/input.go`).
- Tipos de dados de serviços externos (ex: `QuizQuestion`) estiverem definidos dentro dos arquivos de interface de serviço.

### ✅ Verificar:
- `application/commands/` contém apenas structs de entrada para operações de **escrita**.
- `application/queries/` contém apenas structs de entrada para operações de **leitura**.
- `application/dto/` contém todos os DTOs de **resposta** (output).
- `application/mappers/` contém todas as funções de mapeamento `domain ↔ DTO`.
- `ports/commands.go` contém input ports de escrita; `ports/queries.go` contém input ports de leitura.

---

## 4. Interfaces e Ports (ISP / DIP)

### ❌ Rejeitar se:
- Controllers dependerem de structs concretas de use case (ao invés das interfaces em `ports/`).
- Use cases dependerem de structs concretas de repositório ou serviço.
- Interfaces de repositório estiverem fora de `domain/repositories/`.
- Interfaces de serviços externos estiverem fora de `application/services/`.
- Um controller receber uma interface que não usa diretamente (ex: `ITokenService` apenas para repassar ao middleware).
- Uma interface tiver mais de ~5 métodos sem justificativa clara — avaliar segregação.

### ✅ Verificar:
- Todo controller recebe input ports (`ports/`) no construtor.
- Todo use case recebe interfaces (`repositories.IXxx`, `services.IXxx`) no construtor.
- Nenhuma struct concreta de infraestrutura está tipada diretamente em `application/`.
- Interfaces são pequenas e focadas (ISP).

---

## 5. Modelos GORM vs Entidades de Domínio

### ❌ Rejeitar se:
- Entidades de domínio tiverem tags `gorm:` diretamente.
- Repositórios passarem entidades de domínio diretamente ao GORM (sem modelo intermediário).
- Repositórios retornarem modelos GORM ao invés de entidades de domínio.
- A camada de `application/` importar ou conhecer a existência de modelos GORM.
- O nome da tabela for definido via `db.Table("...")` no construtor do repositório ao invés do método `TableName()` do model.

### ✅ Verificar:
- Modelos GORM ficam exclusivamente em `infrastructure/database/models/`.
- Cada model tem o método `TableName() string` definido.
- Repositórios fazem o mapeamento `db.Model → domain.Entity` antes de retornar.
- Funções `ToModel()` e `ToDomain()` existem e estão corretas.

---

## 6. Use Cases (SRP / Clean Code)

### ❌ Rejeitar se:
- Um use case fizer mais de uma responsabilidade de negócio no mesmo `Run`.
- O método principal não seguir o padrão `Run(ctx context.Context, input ...) (*Response, error)`.
- Erros de negócio forem silenciados ou retornados sem `%w` (impossibilitando `errors.Is`).
- Lógica de formatação HTTP (status codes, headers) existir no use case.
- Um use case de escrita não validar invariantes de domínio antes de persistir (ex: `SaveQuizHistory` sem verificar duplicata).
- `panic` for usado para erros de negócio esperados.

### ✅ Verificar:
- Use case tem uma única responsabilidade clara.
- Construtor (`NewXxxUseCase`) recebe todas as dependências como interfaces.
- Erros são propagados com `fmt.Errorf("contexto: %w", err)`.
- `context.Context` é o primeiro parâmetro do `Run`.

---

## 7. Controllers (Clean Code / HTTP Semantics)

### ❌ Rejeitar se:
- Controller acessar repositório ou serviço diretamente (bypass do use case).
- Controller contiver lógica condicional de negócio.
- Erros não forem tratados com o status HTTP **semanticamente correto** (ex: 404 para erro de autorização — deve ser 403).
- Binding de request não usar `ShouldBindJSON` / `ShouldBindQuery`.
- Campos de struct do controller forem exportados (public) sem justificativa.
- Literais inteiros forem usados para status HTTP (ex: `c.JSON(200, ...)` ao invés de `c.JSON(http.StatusOK, ...)`).
- Erros internos (500) exporem mensagens de infraestrutura ao cliente.
- Dois controllers diferentes exporem o mesmo endpoint com a mesma lógica (ex: `POST /api/users` e `POST /auth/register` chamando o mesmo use case).

### ✅ Verificar:
- Fluxo do controller: `bind input → chamar use case via port → formatar resposta`.
- Erros mapeados com `errors.Is()` para tipos semânticos de `domain/errors/`.
- Resposta de erro no formato `{ "message": "..." }`.
- Status HTTP semântico para cada tipo de erro:
  - `400` → validação de input inválido
  - `401` → não autenticado
  - `403` → autenticado mas sem permissão
  - `404` → recurso não encontrado
  - `409` → conflito (duplicata)
  - `422` → entidade não processável (regra de negócio)
  - `500` → erro interno (com mensagem genérica)
- `context.Context` propagado via `c.Request.Context()`.
- Nenhuma variável global de dependência no controller.

---

## 8. Tratamento de Erros (Go Best Practices)

### ❌ Rejeitar se:
- Erros forem ignorados com `_` em operações críticas (DB, hash, token, I/O).
- `panic` for usado para fluxo de controle de erros esperados.
- Strings de erro forem comparadas diretamente (`err.Error() == "..."`) ao invés de `errors.Is()`.
- Erros forem criados sem `%w` impossibilitando `errors.Is()` em camadas superiores.
- Erros de banco/ORM forem expostos diretamente ao cliente HTTP.
- Constructores que podem falhar retornarem apenas `*T` (sem `error`).

### ✅ Verificar:
- Todos os erros de repositórios e serviços são checados e propagados.
- Erros de negócio usam sentinel errors de `domain/errors/`.
- Erros são wrappados com `fmt.Errorf("contexto: %w", err)` para manter a cadeia.
- Checagem de erros usa `errors.Is()` ou `errors.As()`.
- Erros 500 são logados internamente; o cliente recebe apenas `"internal server error"`.

---

## 9. Boas Práticas de Go

### ❌ Rejeitar se:
- `context.Context` não for propagado em operações de I/O (repositórios, serviços externos).
- `context.Context` for armazenado em structs ao invés de passado como parâmetro.
- `bcrypt.MaxCost` for usado (custo 31 — proibitivo; deve ser `bcrypt.DefaultCost`).
- `init()` for usado para inicializar dependências ou lógica de negócio.
- Variáveis globais forem adicionadas para estado ou dependências.
- `log.Fatalf` / `os.Exit` forem chamados fora de `cmd/`.
- `defer` for usado dentro de loops (acumula até o fim da função, não do loop).
- Goroutines forem lançadas sem mecanismo de cancelamento/waitgroup.
- Nomes de constantes não forem semânticos e precisos.
- Package name de `domain/entities/` não for `entities`.

### ✅ Verificar:
- `context.Context` é o primeiro parâmetro em todo método de I/O.
- `c.Request.Context()` é usado nos controllers para propagar o contexto HTTP.
- `bcrypt.DefaultCost` é usado no hasher.
- Constructores que podem falhar retornam `(*T, error)`.
- Nomes de packages são curtos, lowercase, sem underscore.
- Nomes de interfaces seguem o prefixo `I` conforme convenção do projeto.
- Nomes de constantes de tipo são precisos: `DifficultyEasy`, não `DifficultEasy`.
- Imports organizados: stdlib → externos → internos (separados por linha em branco).
- Nenhum campo de struct exportado desnecessariamente.

---

## 10. Segurança

### ❌ Rejeitar se:
- Senhas forem armazenadas ou logadas em texto puro.
- JWT secret, API keys ou qualquer credential estiver hardcoded no código fonte.
- Rotas protegidas não tiverem middleware de autenticação.
- Inputs não forem validados antes de chegarem ao use case (binding com `binding:"required,..."`).
- `GEMINI_API_KEY`, `JWT_SECRET_KEY` ou qualquer secret aparecer no código.
- Detalhes de infraestrutura (mensagens de erro de banco, stack traces) forem expostos ao cliente.

### ✅ Verificar:
- Todas as configurações sensíveis vêm de variáveis de ambiente via `cmd/config/environment.go`.
- Senhas passam por `IHasher.HashPassword` antes de persistir.
- JWT tem `exp` (tempo de expiração) sempre definido.
- Middleware de auth aplicado no grupo de rotas protegidas (configurado na factory).
- Respostas de erro não revelam informações internas.

---

## 11. Testes

### ❌ Rejeitar se:
- Testes unitários instanciarem infraestrutura real (banco, APIs externas).
- Novos use cases forem adicionados sem testes unitários correspondentes.
- Novos controllers forem adicionados sem testes com `httptest`.
- Testes usarem `t.Log` para verificações ao invés de `t.Errorf` / `t.Fatalf`.
- Múltiplos casos de teste estiverem em funções separadas quando poderiam ser table-driven.

### ✅ Verificar:
- Testes unitários em `_test.go` na mesma pasta do código.
- Use cases testados com mocks das interfaces (`IUserRepository`, `IHasher`, etc.).
- Controllers testados com `httptest.NewRecorder()` e `httptest.NewRequest()`.
- Table-driven tests para múltiplos cenários do mesmo comportamento.
- Casos de erro (not found, conflict, forbidden) são testados além do happy path.

---

## 12. Migrações e Banco de Dados

### ❌ Rejeitar se:
- `AutoMigrate` do GORM for usado para criar/alterar tabelas.
- Migrações editarem arquivos já existentes em `db/migrations/`.
- O arquivo de migração não seguir a nomenclatura semântica (`YYYYMMDDHHMMSS_nome_semantico.sql`).
- A migração foi criada manualmente ao invés de usar `make create-migration`.
- Nomenclatura for genérica ou não descritiva (ex: `update_table`, `fix_bug`, `v2`).
- Migrações não forem idempotentes (ausência de `IF NOT EXISTS` / `IF EXISTS`).

### ✅ Verificar:
- Migração criada com `make create-migration nome_semantico`.
- Nomenclatura semântica descreve claramente a operação:
  - ✅ `create_users_table`, `add_email_column_to_users`, `drop_unused_index`
  - ❌ `migration1`, `fix`, `update`, `v2`
- Cada migração é um arquivo novo e imutável.

---

## 13. Qualidade Geral

### ❌ Rejeitar se:
- Código não compilar ou tiver erros de `go vet`.
- Funções tiverem mais de ~50 linhas sem justificativa clara.
- Código comentado (`// old code`) ou `TODO` sem issue linkada estiverem presentes.
- Imports não utilizados ou variáveis declaradas e não usadas existirem.
- Endpoints duplicados com a mesma lógica existirem em controllers diferentes.
- Typos em nomes de campos públicos de struct (ex: `CreatedDAt` ao invés de `CreatedAt`).

### ✅ Verificar:
- `go build ./...` e `go vet ./...` passam sem erros.
- Nomes de funções, tipos e pacotes são claros e seguem `camelCase`/`PascalCase` Go.
- Imports organizados: stdlib → externos → internos.
- Sem código morto ou comentado.

---

## Checklist Rápido de PR

```
[ ] Regra de dependência entre camadas não violada (domain ← application ← infrastructure)
[ ] Entidades de domínio sem tags de infraestrutura (json/gorm/db)
[ ] Erros de negócio em domain/errors/ como sentinel errors
[ ] Commands e Queries em pacotes separados (CQRS)
[ ] Ports de commands e queries em arquivos separados
[ ] Mappers em application/mappers/ — não em commands/ ou queries/
[ ] Modelos GORM em infrastructure/database/models/ com TableName()
[ ] Repositórios mapeiam model → domain (sem expor GORM para application)
[ ] Controllers dependem de ports (interfaces), não de use cases concretos
[ ] Campos de struct do controller são privados (lowercase)
[ ] Controllers usam http.StatusXxx — nunca literais inteiros
[ ] Status HTTP correto para cada tipo de erro (400/401/403/404/409/422/500)
[ ] Erros verificados com errors.Is() / errors.As()
[ ] Erros 500 não expõem detalhes de infraestrutura ao cliente
[ ] context.Context propagado em todo I/O
[ ] log.Fatalf apenas em cmd/
[ ] Constructores que podem falhar retornam (*T, error)
[ ] Middlewares criados nas factories, injetados prontos nos controllers
[ ] bcrypt.DefaultCost (nunca MaxCost)
[ ] Nenhum secret hardcoded
[ ] Migração nova (se houver) segue nomenclatura semântica e é imutável
[ ] Testes unitários com mocks cobrindo happy path e casos de erro
[ ] Table-driven tests para múltiplos cenários
[ ] Código compila e go vet passa sem erros
[ ] Sem typos em nomes de campos públicos
```
