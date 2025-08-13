# Ejemplos de API - Bank Service API

## Autenticación

Todos los endpoints (excepto health checks) requieren un token JWT en el header Authorization:

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Health Checks

### GET /health

**Request:**
```bash
curl -X GET http://localhost:3000/health
```

**Response 200 - Healthy:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "version": "1.0.0",
  "environment": "development"
}
```

**Response 503 - Unhealthy:**
```json
{
  "status": "unhealthy",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "error": "Database connection failed"
}
```

### GET /health/jwt

**Request:**
```bash
curl -X GET http://localhost:3000/health/jwt
```

**Response 200 - JWT Healthy:**
```json
{
  "jwt": {
    "status": "healthy",
    "secretProvider": "AWSSecretsProvider"
  },
  "service": "healthy",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

## Bancos

### GET /api/banks - Listar Bancos

**Request básico:**
```bash
curl -X GET "http://localhost:3000/api/banks" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Request con filtros:**
```bash
curl -X GET "http://localhost:3000/api/banks?env=production&country=ES&page=1&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "bank_id": "santander_es",
      "name": "Banco Santander España",
      "bank_codes": ["0049"],
      "api": "berlin_group",
      "api_version": "1.3.6",
      "aspsp": "santander",
      "country": "ES",
      "auth_type_choice_required": true,
      "bic": "BSCHESMM",
      "real_name": "Banco Santander S.A.",
      "product_code": null,
      "bank_group_id": "santander_group",
      "logo_url": "https://cdn.banks.com/logos/santander_es.png",
      "documentation": "API documentation available at...",
      "keywords": {
        "popular": true,
        "recommended": true
      },
      "attribute": {
        "supports_psd2": true,
        "instant_payments": true
      }
    },
    {
      "bank_id": "bbva_es",
      "name": "BBVA España",
      "bank_codes": ["0182"],
      "api": "berlin_group",
      "api_version": "1.3.6",
      "aspsp": "bbva",
      "country": "ES",
      "auth_type_choice_required": false,
      "bic": "BBVAESMM",
      "real_name": "Banco Bilbao Vizcaya Argentaria S.A.",
      "product_code": null,
      "bank_group_id": "bbva_group",
      "logo_url": "https://cdn.banks.com/logos/bbva_es.png",
      "documentation": null,
      "keywords": {
        "popular": true
      },
      "attribute": {
        "supports_psd2": true
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 145,
    "totalPages": 15
  }
}
```

### GET /api/banks/{bankId}/details - Detalles de Banco

**Request:**
```bash
curl -X GET "http://localhost:3000/api/banks/santander_es/details" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "bank_id": "santander_es",
    "name": "Banco Santander España",
    "bank_codes": ["0049"],
    "api": "berlin_group",
    "api_version": "1.3.6",
    "aspsp": "santander",
    "country": "ES",
    "auth_type_choice_required": true,
    "bic": "BSCHESMM",
    "real_name": "Banco Santander S.A.",
    "product_code": null,
    "bank_group_id": "santander_group",
    "logo_url": "https://cdn.banks.com/logos/santander_es.png",
    "documentation": "API documentation available at...",
    "keywords": {
      "popular": true,
      "recommended": true
    },
    "attribute": {
      "supports_psd2": true,
      "instant_payments": true
    },
    "environment_configs": [
      {
        "environment": "production",
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": null,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 1000,
        "ok_status_codes_simple_payment": {
          "200": "OK",
          "201": "Created"
        },
        "ok_status_codes_instant_payment": {
          "200": "OK",
          "201": "Created"
        },
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": false,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      },
      {
        "environment": "test",
        "enabled": 1,
        "blocked": false,
        "risky": true,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": "Este es un ambiente de pruebas",
        "supports_instant_payments": true,
        "instant_payments_activated": false,
        "instant_payments_limit": 100,
        "ok_status_codes_simple_payment": {
          "200": "OK"
        },
        "ok_status_codes_instant_payment": null,
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": false,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      }
    ]
  }
}
```

**Response 404 - Bank Not Found:**
```json
{
  "success": false,
  "error": "Bank not found"
}
```

### POST /api/banks - Crear Banco

**Request - Con array de ambientes:**
```bash
curl -X POST "http://localhost:3000/api/banks" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "bank_id": "caixabank_es",
    "name": "CaixaBank",
    "bank_codes": ["2100"],
    "api": "berlin_group",
    "api_version": "1.3.6",
    "aspsp": "caixabank",
    "country": "ES",
    "auth_type_choice_required": true,
    "bic": "CAIXESBB",
    "real_name": "CaixaBank S.A.",
    "bank_group_id": "caixa_group",
    "logo_url": "https://cdn.banks.com/logos/caixabank_es.png",
    "keywords": {
      "popular": true
    },
    "environments": ["production", "test"],
    "configuration": {
      "enabled": 1,
      "blocked": false,
      "risky": false,
      "app_auth_setup_required": false,
      "supports_instant_payments": true,
      "instant_payments_activated": true,
      "instant_payments_limit": 1500
    }
  }'
```

**Request - Con configuraciones específicas:**
```bash
curl -X POST "http://localhost:3000/api/banks" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "bank_id": "ing_es",
    "name": "ING España",
    "bank_codes": ["1465"],
    "api": "berlin_group",
    "api_version": "1.3.6",
    "aspsp": "ing",
    "country": "ES",
    "auth_type_choice_required": false,
    "bic": "INGDESMMXXX",
    "configurations": {
      "production": {
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": true,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 2000
      },
      "test": {
        "enabled": 1,
        "blocked": false,
        "risky": true,
        "app_auth_setup_required": false,
        "supports_instant_payments": false,
        "risky_message": "Ambiente de pruebas - funcionalidad limitada"
      }
    }
  }'
```

**Response 201 - Created:**
```json
{
  "success": true,
  "data": {
    "bank": {
      "bank_id": "caixabank_es",
      "name": "CaixaBank",
      "bank_codes": ["2100"],
      "api": "berlin_group",
      "api_version": "1.3.6",
      "aspsp": "caixabank",
      "country": "ES",
      "auth_type_choice_required": true,
      "bic": "CAIXESBB",
      "real_name": "CaixaBank S.A.",
      "product_code": null,
      "bank_group_id": "caixa_group",
      "logo_url": "https://cdn.banks.com/logos/caixabank_es.png",
      "documentation": null,
      "keywords": {
        "popular": true
      },
      "attribute": null
    },
    "environment_configs": [
      {
        "environment": "production",
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": null,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 1500,
        "ok_status_codes_simple_payment": null,
        "ok_status_codes_instant_payment": null,
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": null,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      },
      {
        "environment": "test",
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": null,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 1500,
        "ok_status_codes_simple_payment": null,
        "ok_status_codes_instant_payment": null,
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": null,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      }
    ]
  },
  "message": "Bank created successfully"
}
```

### PUT /api/banks/{bankId} - Actualizar Banco

**Request:**
```bash
curl -X PUT "http://localhost:3000/api/banks/caixabank_es" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CaixaBank España (Actualizado)",
    "api_version": "1.4.0",
    "logo_url": "https://cdn.banks.com/logos/caixabank_es_new.png",
    "configurations": {
      "production": {
        "instant_payments_limit": 2000,
        "ok_status_codes_simple_payment": {
          "200": "OK",
          "201": "Created",
          "202": "Accepted"
        }
      },
      "test": {
        "risky": false,
        "risky_message": null
      }
    }
  }'
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "bank": {
      "bank_id": "caixabank_es",
      "name": "CaixaBank España (Actualizado)",
      "bank_codes": ["2100"],
      "api": "berlin_group",
      "api_version": "1.4.0",
      "aspsp": "caixabank",
      "country": "ES",
      "auth_type_choice_required": true,
      "bic": "CAIXESBB",
      "real_name": "CaixaBank S.A.",
      "product_code": null,
      "bank_group_id": "caixa_group",
      "logo_url": "https://cdn.banks.com/logos/caixabank_es_new.png",
      "documentation": null,
      "keywords": {
        "popular": true
      },
      "attribute": null
    },
    "environment_configs": [
      {
        "environment": "production",
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": null,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 2000,
        "ok_status_codes_simple_payment": {
          "200": "OK",
          "201": "Created",
          "202": "Accepted"
        },
        "ok_status_codes_instant_payment": null,
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": null,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      },
      {
        "environment": "test",
        "enabled": 1,
        "blocked": false,
        "risky": false,
        "app_auth_setup_required": false,
        "blocked_text": null,
        "risky_message": null,
        "supports_instant_payments": true,
        "instant_payments_activated": true,
        "instant_payments_limit": 1500,
        "ok_status_codes_simple_payment": null,
        "ok_status_codes_instant_payment": null,
        "ok_status_codes_periodic_payment": null,
        "enabled_periodic_payment": null,
        "frequency_periodic_payment": null,
        "config_periodic_payment": null
      }
    ]
  },
  "message": "Bank updated successfully"
}
```

## Grupos Bancarios

### GET /api/bank-groups - Listar Grupos Bancarios

**Request:**
```bash
curl -X GET "http://localhost:3000/api/bank-groups" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "group_id": "santander_group",
      "name": "Grupo Santander",
      "description": "Uno de los principales grupos bancarios del mundo",
      "logo_url": "https://cdn.banks.com/logos/santander_group.png",
      "website": "https://www.santander.com",
      "created_at": "2024-01-01T00:00:00.000Z",
      "updated_at": "2024-01-15T10:30:00.000Z"
    },
    {
      "group_id": "bbva_group",
      "name": "Grupo BBVA",
      "description": "Grupo financiero global",
      "logo_url": "https://cdn.banks.com/logos/bbva_group.png",
      "website": "https://www.bbva.com",
      "created_at": "2024-01-01T00:00:00.000Z",
      "updated_at": "2024-01-10T15:20:00.000Z"
    }
  ]
}
```

### GET /api/bank-groups/{groupId} - Detalles de Grupo Bancario

**Request:**
```bash
curl -X GET "http://localhost:3000/api/bank-groups/santander_group" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "group_id": "santander_group",
    "name": "Grupo Santander",
    "description": "Uno de los principales grupos bancarios del mundo",
    "logo_url": "https://cdn.banks.com/logos/santander_group.png",
    "website": "https://www.santander.com",
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-15T10:30:00.000Z"
  }
}
```

## Filtros

### GET /api/filters - Obtener Filtros

**Request:**
```bash
curl -X GET "http://localhost:3000/api/filters" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "countries": [
      {
        "code": "ES",
        "name": "España",
        "count": 45
      },
      {
        "code": "FR",
        "name": "Francia",
        "count": 38
      },
      {
        "code": "DE",
        "name": "Alemania",
        "count": 52
      }
    ],
    "apis": [
      {
        "type": "berlin_group",
        "count": 120
      },
      {
        "type": "openbanking_uk",
        "count": 85
      },
      {
        "type": "custom",
        "count": 45
      }
    ],
    "environments": [
      "production",
      "test",
      "sandbox",
      "development"
    ],
    "bankGroups": [
      {
        "group_id": "santander_group",
        "name": "Grupo Santander",
        "count": 8
      },
      {
        "group_id": "bbva_group",
        "name": "Grupo BBVA",
        "count": 6
      }
    ]
  }
}
```

## Errores Comunes

### 400 - Bad Request

**Ejemplo - Parámetros inválidos:**
```json
{
  "success": false,
  "error": "Invalid environment parameter. Must be one of: production, test, sandbox, development, all"
}
```

**Ejemplo - Validación fallida:**
```json
{
  "success": false,
  "error": "Validation failed: bank_id is required"
}
```

### 401 - Unauthorized

**Token faltante:**
```json
{
  "success": false,
  "error": "Authentication required"
}
```

**Token inválido:**
```json
{
  "success": false,
  "error": "Invalid JWT token"
}
```

**Token expirado:**
```json
{
  "success": false,
  "error": "Token expired"
}
```

### 403 - Forbidden

**Permisos insuficientes:**
```json
{
  "success": false,
  "error": "Missing required permission: banks:write"
}
```

### 404 - Not Found

**Endpoint no encontrado:**
```json
{
  "success": false,
  "error": "Endpoint not found",
  "path": "/api/invalid-endpoint",
  "method": "GET"
}
```

**Recurso no encontrado:**
```json
{
  "success": false,
  "error": "Bank not found"
}
```

### 409 - Conflict

**Recurso ya existe:**
```json
{
  "success": false,
  "error": "Bank with ID 'santander_es' already exists"
}
```

### 500 - Internal Server Error

**Error de desarrollo:**
```json
{
  "success": false,
  "error": "Database connection failed",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

**Error de producción:**
```json
{
  "success": false,
  "error": "Internal server error",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

## Rate Limiting

Cuando se excede el límite de requests:

**Response 429:**
```json
{
  "success": false,
  "error": "Too many requests"
}
```

**Headers incluidos:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1642248600
Retry-After: 900
```