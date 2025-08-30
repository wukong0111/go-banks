## Visión General del Proyecto

Proyecto Go para API de servicio bancario. Usa Gin para framework web, PostgreSQL como base de datos, JWT para autenticación. Contenerizado con Docker y gestionado por Makefile.

**Tecnologías Clave:**
- Go: Lenguaje principal.
- Gin: Framework para APIs.
- PostgreSQL: Base de datos relacional.
- Docker: Contenerización.
- JWT: Seguridad con tokens.
- golang-migrate: Migraciones de DB.

**Arquitectura:**
- `cmd/`: Aplicaciones principales (API, migraciones, seeding).
- `internal/`: Lógica de negocio (handlers, models, repositories, services).
- `migrations/`: Archivos SQL de migraciones.
- `seeders/`: Archivos SQL para datos iniciales.

## Construcción y Ejecución

**Prerrequisitos:**
- Go 1.25.0 (requerido para features modernas).
- Docker y Docker Compose.
- `make`.

**Comandos Clave:**
- `make dev`: Inicia entorno dev con live reload.
- `make build`: Compila y crea binario en `bin/`.
- `make run`: Ejecuta la app compilada.
- `make test`: Ejecuta tests.
- Migraciones: `make migrate-up` (aplicar), `make migrate-down` (revertir), `make migrate-status` (estado).
- Seeding: `make seed`.

## Convenciones de Desarrollo

### Estándares de Código

**Versión Requerida:** Go 1.25.0 - Usa features modernas:
- Range over integers: `for i := range 10 { ... }`.
- Inferencia de tipos mejorada.
- Iteradores mejorados.
- Sintaxis moderna: `any` en vez de `interface{}`.
- Paquete slices: `slices.Contains()` en vez de loops manuales.

### Guías de Seguridad de Tipos

**Principio:** Maximiza seguridad de tipos con sintaxis moderna.

1. **Modernización de Sintaxis:**
   - Usa `any` no `interface{}`.

2. **Optimización de Seguridad:**
   - Prefiere interfaces específicas para polimorfismo.
   - Usa `any` cuando necesario, con sintaxis moderna.

3. **Tipos Genéricos:**
   - Para reusabilidad: `type APIResponse[T any] struct { ... }`.
   - Alternativa: `any` con type switches.

**Casos Aceptables para `any`:**
- Unmarshaling JSON: `map[string]any`.
- Campos JSONB en DB.
- Interfaces de libs externas.
- Contenedores genéricos con constraints.

### Patrones de Código

1. **Manejo de Errores:**
   - Patrón run(): `func main() { if err := run(); err != nil { logger.Error("failed to run application", "error", err) } }`.
   - Usa `defer` para cleanup.

2. **Operaciones con Slices:**
   - Usa paquete slices.
   - Range over integers.

3. **Operaciones con Strings:**
   - Concatenación simple.
   - `errors.New` para mensajes estáticos.
   - `fmt.Errorf` para formateo complejo.

4. **Patrones HTTP:**
   - `http.NoBody` para requests vacíos.
   - Manejo de errores con switch en strings.

### Estándares de Testing

**Framework:** testify.
- `require` para assertions críticas.
- `assert` para checks de valores.
- `mock` para mocking.

### Configuración de Linter

**golangci-lint** con estándares estrictos - 0 issues tolerados.

**Linters Habilitados:**
- gosec, gocritic, staticcheck, revive, perfsprint, prealloc, ineffassign, unused, misspell.

**Ejecución:**
- `make lint`, `make test`, `make build` deben pasar limpios.

**Política Crítica:**
- Nunca suprime warnings editando config o excluyendo archivos.
- Siempre fija el código subyacente.
- Cambios aceptables: Ajustes de thresholds legítimos, nuevos linters.

## Conocimiento de Dominio

### Conceptos Bancarios

**Entornos:**
- `sandbox`: Dev/testing.
- `production`: Operaciones live.
- `uat`: Testing de aceptación.
- `test`: Testing automatizado.

**Modelos Core:**
- `Bank`: Institución con país, tipo API, entornos.
- `BankGroup`: Grupo de bancos relacionados.
- `BankEnvironmentConfig`: Config por entorno.
- `BankDetails`: Interfaz polimórfica (single/multiple entornos).

**Patrones de Datos:**
- Campos JSONB: `map[string]any`.
- Arrays: `pgtype.Array[string]`.
- Respuestas: `APIResponse[T any]`.

## Principios de Diseño API

**Especificación:** OpenAPI 3.0.3 en `docs/api-documentation.yml`.

**Autenticación:** JWT requerido (excepto health checks).
- Permisos: `banks:read/write`.
- Validación via middleware.
- Claims: subject, permissions array.

**Patrones de Respuesta:**
- Wrapper `APIResponse[T]`.
- Códigos HTTP apropiados.
- Errores estructurados.
- Paginación en listas.

## Notas Importantes

**Workflows Críticos:**
- Siempre migraciones antes de seeding: `make migrate-up && make seed`.
- Nunca commit con issues: `make lint` debe mostrar 0 issues, `make test` pasar.
- Test antes de commit.

**Específicos de Entorno:**
- Usa `docker compose` (no `docker-compose`).
- PostgreSQL via container Docker.
- Go 1.25.0 requerido para features modernas.

**Operaciones DB:**
- Migraciones versionadas (001-004).
- Seeds dependen de migraciones actuales.
- `make db-reset` para reset dev.

**Operaciones Server:**
- Shutdown graceful: Maneja SIGTERM/SIGINT.
- Health checks: `/health` (liveness), `/ready` (readiness con check DB).
- Timeouts: 15s read/write, 60s idle, 30s shutdown.
- Listo para producción: Compatible con Kubernetes, Docker Swarm.
