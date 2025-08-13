# Autenticación JWT - Bank Service API

## Descripción General

La API utiliza un sistema de autenticación JWT simplificado que elimina la complejidad de la gestión individual de claves API. En lugar de manejar cientos de claves API, utiliza un secreto JWT compartido almacenado en AWS Secrets Manager.

## Características Principales

- **Secreto Compartido Único**: Un secreto JWT por ambiente almacenado en AWS Secrets Manager
- **Validación Sin Estado**: No requiere consultas a base de datos para validar tokens
- **Expiración Automática**: Los tokens expiran automáticamente sin revocación manual
- **Rotación Automática**: Rotación proactiva de secretos con notificaciones SNS/SQS
- **Diseño Basado en Interfaces**: Múltiples implementaciones para producción, desarrollo y testing

## Estructura de Claims JWT

### Claims Estándar (RFC 7519)

```json
{
  "iss": "bank-api-client",           // Identificador del servicio emisor
  "sub": "service-auth",              // Sujeto (tipo de autenticación)
  "exp": 1234567890,                  // Timestamp de expiración
  "iat": 1234567800                   // Timestamp de emisión
}
```

### Claims de Negocio

```json
{
  "service_type": "internal",         // Tipo de servicio
  "permissions": [                    // Permisos otorgados
    "banks:read",
    "banks:write"
  ],
  "environment": "production"         // Ambiente de ejecución
}
```

### Claims Opcionales

```json
{
  "version": "1.0.0",                // Versión del servicio
  "region": "us-east-1"              // Región AWS
}
```

### Ejemplo Completo de Token JWT

```json
{
  "iss": "bank-api-client",
  "sub": "service-auth",
  "exp": 1734567890,
  "iat": 1734567800,
  "service_type": "internal",
  "permissions": ["banks:read", "banks:write"],
  "environment": "production",
  "version": "1.0.0",
  "region": "us-east-1"
}
```

## Tipos de Servicio

| service_type | Descripción | Permisos Típicos |
|-------------|-------------|------------------|
| `internal` | Servicios internos de la organización | `banks:read`, `banks:write` |
| `external` | Servicios de terceros con acceso limitado | `banks:read` |
| `partner` | Socios comerciales con permisos específicos | `banks:read`, permisos específicos |

## Permisos Disponibles

### Permisos de Bancos
- `banks:read` - Lectura de datos de bancos y filtros
- `banks:write` - Creación y actualización de bancos

### Permisos de Grupos Bancarios
- `bank-groups:read` - Lectura de grupos bancarios (incluido en `banks:read`)

## Ambientes Soportados

| Ambiente | Descripción | Configuración |
|----------|-------------|---------------|
| `production` | Ambiente de producción | Secretos AWS, rotación cada 30 días |
| `test` | Ambiente de testing/QA | Secretos AWS, rotación cada 7 días |
| `sandbox` | Ambiente de desarrollo compartido | Secretos AWS, rotación manual |
| `development` | Ambiente local | Variables de entorno, sin rotación |


## Rotación de Secretos

### Arquitectura de Rotación

El sistema utiliza una arquitectura orientada a eventos para rotación proactiva:

```
┌─────────────────────────────────────────────────────────────┐
│                    Flujo de Rotación                        │
├─────────────────────────────────────────────────────────────┤
│ AWS Lambda (Rotation Trigger)                              │
│     ↓                                                       │
│ AWS Secrets Manager (Rotate Secret)                        │
│     ↓                                                       │
│ SNS Topic (jwt-secret-rotation)                           │
│     ↓                                                       │
│ SQS Queue (jwt-secret-rotation-queue)                     │
│     ↓                                                       │
│ RotationOrchestrator                                       │
│ ├─ SimpleJWTService (cache invalidation)                  │
│ └─ Event listeners por ambiente                           │
└─────────────────────────────────────────────────────────────┘
```

## Seguridad

### Características de Seguridad

- **Longitud mínima de secreto**: 256 bits
- **Algoritmo**: HS256 (HMAC SHA-256)
- **Período de gracia**: 5 minutos durante rotación
- **Aislamiento por ambiente**: Secretos separados por ambiente
- **Logging de auditoría**: Registro completo de validaciones

## Desarrollo Local

### Variables de Entorno Requeridas

```bash
# AWS (Producción)
AWS_REGION=us-east-1
AWS_SECRETS_MANAGER_SECRET_NAME=jwt-secret-prod

# Local (Desarrollo)
JWT_SECRET_LOCAL=your-256-bit-secret-key-here
NODE_ENV=development

# SNS/SQS (Eventos de rotación)
SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789:jwt-secret-rotation
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/jwt-secret-rotation-queue
```

### Testing con LocalStack

```bash
# Iniciar LocalStack con SNS/SQS
make dev-localstack

# Configurar recursos AWS
make aws-setup

# Probar eventos de rotación
make aws-test-rotation
```
