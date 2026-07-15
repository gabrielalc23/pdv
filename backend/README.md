# Backend

Backend Go do PDV com PostgreSQL, `pgx/v5`, `sqlc` e servidor HTTP.

## Execução

```sh
cd backend
go run ./cmd/api
```

Variáveis de ambiente:

- `DATABASE_URL` obrigatória;
- `HTTP_ADDRESS` opcional, padrão `:8080`.

## Produtos

### Endpoints

- `POST /products`
- `GET /products`
- `GET /products/{id}`
- `PUT /products/{id}`
- `POST /products/{id}/activate`
- `POST /products/{id}/deactivate`

### Listagem

`GET /products` aceita:

- `search`
- `page`
- `pageSize`
- `activeOnly`

Exemplo:

```sh
curl 'http://localhost:8080/products?search=coca&page=1&pageSize=20&activeOnly=true'
```

### Criação e atualização

Exemplo de request:

```json
{
  "sku": "COCA-2L",
  "barcode": "7890000000000",
  "name": "Coca-Cola 2L",
  "price": "12.90",
  "cost": "8.50"
}
```

Exemplo de response:

```json
{
  "id": "01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9",
  "sku": "COCA-2L",
  "barcode": "7890000000000",
  "name": "Coca-Cola 2L",
  "price": "12.90",
  "cost": "8.50",
  "isActive": true,
  "createdAt": "2026-07-15T10:00:00Z",
  "updatedAt": "2026-07-15T10:00:00Z"
}
```

### Principais erros

- `400 Bad Request` para JSON inválido, parâmetros inválidos ou `id` inválido;
- `404 Not Found` para produto inexistente;
- `409 Conflict` para SKU ou barcode duplicado;
- `422 Unprocessable Entity` para validação semântica do payload;
- `500 Internal Server Error` para falhas inesperadas.

## Validação

Os comandos principais do backend são:

```sh
sqlc generate
sqlc vet
go fmt ./...
go vet ./...
go test ./...
```
