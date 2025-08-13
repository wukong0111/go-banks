# Manejo de Errores - Bank Service API

## Estructura de Respuestas de Error

Todas las respuestas de error siguen el formato estándar `ApiResponse`:

```typescript
interface ErrorResponse {
  success: false;
  error: string;
  timestamp?: string;  // Solo en errores 5xx
  path?: string;       // Solo en errores 404
  method?: string;     // Solo en errores 404
}
```

## Códigos de Estado HTTP

### 2xx - Éxito

| Código | Descripción | Uso |
|--------|-------------|-----|
| `200` | OK | Operaciones GET exitosas, PUT exitosos |
| `201` | Created | POST exitoso (banco creado) |

### 4xx - Errores del Cliente

#### 400 - Bad Request
**Causa**: Parámetros inválidos, datos malformados o validación fallida

**Ejemplos:**
```json
// Parámetro de query inválido
{
  "success": false,
  "error": "Invalid environment parameter. Must be one of: production, test, sandbox, development, all"
}

// Validación de esquema fallida
{
  "success": false,
  "error": "Validation failed: bank_id is required"
}

// Formato de datos incorrecto
{
  "success": false,
  "error": "Invalid JSON format in request body"
}

// Tipo de datos incorrecto
{
  "success": false,
  "error": "bank_codes must be an array of strings"
}
```

**Cuándo se produce:**
- Parámetros de query con valores inválidos
- Body de request con JSON malformado
- Campos requeridos faltantes
- Tipos de datos incorrectos
- Longitud de cadenas fuera de límites

#### 401 - Unauthorized
**Causa**: Autenticación requerida pero no proporcionada o inválida

**Ejemplos:**
```json
// Token faltante
{
  "success": false,
  "error": "Authentication required"
}

// Token con formato inválido
{
  "success": false,
  "error": "Invalid JWT token format"
}

// Token expirado
{
  "success": false,
  "error": "Token expired"
}

// Token con firma inválida
{
  "success": false,
  "error": "Invalid token signature"
}

// Secreto JWT no disponible
{
  "success": false,
  "error": "Unable to validate token: JWT secret not available"
}
```

**Cuándo se produce:**
- Header `Authorization` faltante
- Token JWT malformado
- Token firmado con secreto incorrecto
- Token expirado (`exp` claim)
- Problemas de conectividad con AWS Secrets Manager

#### 403 - Forbidden
**Causa**: Token válido pero sin permisos suficientes

**Ejemplos:**
```json
// Permiso específico faltante
{
  "success": false,
  "error": "Missing required permission: banks:write"
}

// Múltiples permisos faltantes
{
  "success": false,
  "error": "Missing required permissions: banks:read, banks:write"
}

// Ambiente no autorizado
{
  "success": false,
  "error": "Access denied for environment: production"
}

// Tipo de servicio no autorizado
{
  "success": false,
  "error": "Service type 'external' not authorized for this operation"
}
```

**Cuándo se produce:**
- Token válido pero sin el permiso requerido
- Ambiente en el token no coincide con el recurso
- Tipo de servicio con restricciones de acceso

#### 404 - Not Found
**Causa**: Recurso o endpoint no encontrado

**Ejemplos:**
```json
// Endpoint inexistente
{
  "success": false,
  "error": "Endpoint not found",
  "path": "/api/invalid-endpoint",
  "method": "GET"
}

// Banco no encontrado
{
  "success": false,
  "error": "Bank not found"
}

// Grupo bancario no encontrado
{
  "success": false,
  "error": "Bank group not found"
}
```

**Cuándo se produce:**
- URL de endpoint incorrecto
- ID de banco inexistente
- ID de grupo bancario inexistente

#### 409 - Conflict
**Causa**: Conflicto con el estado actual del recurso

**Ejemplos:**
```json
// Banco ya existe
{
  "success": false,
  "error": "Bank with ID 'santander_es' already exists"
}

// Violación de constrains únicos
{
  "success": false,
  "error": "Bank code '0049' already assigned to another bank"
}
```

**Cuándo se produce:**
- Intento de crear banco con ID duplicado
- Violación de restricciones de unicidad en BD

#### 429 - Too Many Requests
**Causa**: Límite de rate limiting excedido

**Ejemplo:**
```json
{
  "success": false,
  "error": "Too many requests"
}
```

**Headers adicionales:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1642248600
Retry-After: 900
```

**Configuración por defecto:**
- **Ventana**: 15 minutos
- **Límite**: 100 requests por ventana
- **Aplicación**: Por IP address

### 5xx - Errores del Servidor

#### 500 - Internal Server Error
**Causa**: Error interno no manejado

**Ejemplos:**
```json
// Error en desarrollo (detallado)
{
  "success": false,
  "error": "Database connection timeout",
  "timestamp": "2024-01-15T10:30:00.000Z"
}

// Error en producción (genérico)
{
  "success": false,
  "error": "Internal server error",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

**Cuándo se produce:**
- Excepción no capturada en el código
- Error de conexión a base de datos
- Timeout en operaciones externas
- Problemas de memoria o recursos

#### 503 - Service Unavailable
**Causa**: Servicio temporalmente no disponible

**Ejemplos:**
```json
// Health check fallido
{
  "status": "unhealthy",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "error": "Database connection failed"
}

// JWT service no disponible
{
  "jwt": {
    "status": "unhealthy",
    "error": "AWS Secrets Manager connection failed"
  },
  "service": "failed",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

**Cuándo se produce:**
- Base de datos inaccesible
- AWS Secrets Manager inaccesible
- Dependencias externas caídas

## Manejo de Errores por Tipo

### Errores de Validación (400)

**Origen**: Middleware de validación (`simplifiedValidator.ts`)

**Estructura típica:**
```json
{
  "success": false,
  "error": "Validation failed: [campo] [motivo]"
}
```

**Campos validados:**
- **bank_id**: Requerido, string no vacío
- **name**: Requerido, string no vacío
- **bank_codes**: Requerido, array de strings
- **api**: Requerido, string no vacío
- **country**: Requerido, código ISO de país
- **environments**: Array de ambientes válidos
- **configurations**: Objeto con configuraciones válidas

### Errores de Autenticación (401/403)

**Origen**: Middleware JWT (`simpleJWTMiddleware.ts`)

**Flujo de validación:**
1. **Verificar header**: `Authorization: Bearer <token>`
2. **Verificar formato**: Token JWT bien formado
3. **Verificar firma**: Validar con secreto actual
4. **Verificar expiración**: Claim `exp` válido
5. **Verificar permisos**: Claims `permissions` suficientes

### Errores de Base de Datos (500)

**Tipos comunes:**
- **Connection timeout**: Timeout en conexión a PostgreSQL
- **Query error**: Error en sintaxis SQL o constraint violations
- **Pool exhausted**: Conexiones del pool agotadas

**Ejemplo de manejo:**
```typescript
try {
  const result = await db.query(sql, params);
  return result.rows;
} catch (error) {
  logger.error('Database query failed', error);
  throw new Error('Database operation failed');
}
```

### Errores de Secrets Manager (401/503)

**Tipos comunes:**
- **Secret not found**: Secreto no existe en AWS
- **Access denied**: Permisos IAM insuficientes
- **Network timeout**: Timeout en conexión a AWS
- **Invalid secret format**: Secreto no es un string válido

**Configuración de retry:**
```typescript
const secretsConfig = {
  maxRetries: 3,
  retryDelayMs: 1000,
  timeoutMs: 5000
};
```

## Logging de Errores

### Estructura de Logs

```typescript
// Info level - Request/Response normal
logger.info('GET /api/banks - 200 (45ms)', {
  method: 'GET',
  path: '/api/banks',
  status: 200,
  duration: 45,
  ip: '192.168.1.1'
});

// Error level - Errores del servidor
logger.error('Database query failed', error, {
  query: 'SELECT * FROM banks WHERE...',
  params: ['santander_es'],
  duration: 5000
});

// Warn level - Errores del cliente
logger.warn('Invalid token provided', {
  ip: '192.168.1.1',
  userAgent: 'curl/7.68.0',
  tokenPrefix: 'eyJhbGc...'
});
```

### Configuración por Ambiente

```typescript
// Development: Logs detallados
{
  level: 'debug',
  includeStackTrace: true,
  includeRequestBody: true
}

// Production: Logs básicos
{
  level: 'info',
  includeStackTrace: false,
  includeRequestBody: false,
  sanitizeHeaders: ['authorization']
}
```

## Monitoreo y Alertas

### Métricas Clave

- **Error rate**: Porcentaje de requests 4xx/5xx
- **Response time P95**: Percentil 95 de tiempo de respuesta
- **Database connection pool**: Conexiones activas/disponibles
- **JWT validation failures**: Rate de fallos de autenticación

### Alertas Configuradas

1. **Error rate > 5%** durante 5 minutos
2. **Response time P95 > 1000ms** durante 3 minutos
3. **Database connections > 80%** del pool
4. **JWT validation failures > 10%** durante 2 minutos

### Health Checks

```bash
# Health check básico (incluye DB)
curl http://localhost:3000/health

# Health check JWT (incluye Secrets Manager)
curl http://localhost:3000/health/jwt
```

## Mejores Prácticas

### Para Clientes de la API

1. **Siempre verificar `success` field**:
   ```typescript
   if (response.success) {
     // Usar response.data
   } else {
     // Manejar response.error
   }
   ```

2. **Implementar retry con backoff** para errores 5xx:
   ```typescript
   const retryDelays = [1000, 2000, 4000]; // 1s, 2s, 4s
   ```

3. **Cachear tokens JWT** hasta poco antes de expiración:
   ```typescript
   // Renovar token 5 minutos antes de expiración
   const renewalBuffer = 5 * 60 * 1000;
   ```

4. **Manejar rate limiting** apropiadamente:
   ```typescript
   if (response.status === 429) {
     const retryAfter = response.headers['retry-after'];
     await delay(retryAfter * 1000);
   }
   ```

### Para Desarrolladores

1. **Usar códigos de estado apropiados**
2. **Incluir mensajes de error descriptivos** 
3. **Log errors con contexto suficiente**
4. **Implementar circuit breakers** para dependencias externas
5. **Validar entrada en múltiples capas**

## Troubleshooting Común

### "Token expired"
1. Generar nuevo token: `npm run jwt:generate`
2. Verificar sincronización de relojes
3. Ajustar `expiresInHours` si es necesario

### "Database connection failed"
1. Verificar configuración de BD: `make logs`
2. Reiniciar contenedores: `make down && make up`
3. Verificar logs de PostgreSQL

### "Missing required permission"
1. Verificar claims del token: decode JWT en jwt.io
2. Generar token con permisos correctos
3. Verificar configuración de middleware

### Rate limiting inesperado
1. Verificar múltiples IPs detrás de proxy
2. Ajustar configuración de rate limiting
3. Implementar whitelist para IPs internas