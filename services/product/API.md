# Product Service API Specification

**Service Name:** product-service  
**Port:** 8081  
**Base URL:** `http://localhost:8081`

---

## Overview

Product Service cung cấp các API để quản lý sản phẩm (CRUD operations). Tất cả dữ liệu được lưu trữ trong database MySQL.

---

## Endpoints

### 1. Health Check

**GET** `/health`

Kiểm tra trạng thái của service.

**Response:** `200 OK`
```json
{
  "status": "ok"
}
```

**Postman Test:**
```
GET http://localhost:8081/health
```

---

### 2. List All Products

**GET** `/products`

Lấy danh sách tất cả sản phẩm từ database.

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Basic T-shirt",
    "price": 19.99,
    "category_id": 1,
    "created_at": "2026-04-18T09:20:00Z"
  },
  {
    "id": 2,
    "name": "Sneakers",
    "price": 59.99,
    "category_id": 2,
    "created_at": "2026-04-17T10:10:00Z"
  },
  {
    "id": 3,
    "name": "Coffee Mug",
    "price": 9.99,
    "category_id": 3,
    "created_at": "2026-04-16T15:05:00Z"
  }
]
```

**Postman Test:**
```
GET http://localhost:8081/products
```

---

### 3. Get Product by ID

**GET** `/products/{id}`

Lấy thông tin chi tiết của một sản phẩm theo ID.

**Path Parameters:**
- `id` (int64) - ID của sản phẩm

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Basic T-shirt",
  "price": 19.99,
  "category_id": 1,
  "created_at": "2026-04-18T09:20:00Z"
}
```

**Error Response:** `404 Not Found`
```
product not found
```

**Postman Test:**
```
GET http://localhost:8081/products/1
GET http://localhost:8081/products/2
GET http://localhost:8081/products/999  (Not Found)
```

---

### 4. Create New Product

**POST** `/products`

Tạo sản phẩm mới.

**Request Body:**
```json
{
  "name": "New T-shirt",
  "price": 24.99,
  "category_id": 1
}
```

**Required Fields:**
- `name` (string) - Tên sản phẩm
- `price` (float64) - Giá sản phẩm (phải > 0)
- `category_id` (int64) - ID danh mục

**Response:** `201 Created`
```json
{
  "id": 4,
  "name": "New T-shirt",
  "price": 24.99,
  "category_id": 1,
  "created_at": "2026-04-22T10:30:00Z"
}
```

**Error Response:** `400 Bad Request`
```
name and price are required and price must be positive
```

**Postman Test:**
```
POST http://localhost:8081/products
Content-Type: application/json

{
  "name": "Premium Hoodie",
  "price": 49.99,
  "category_id": 1
}
```

---

### 5. Update Product

**PUT** `/products/{id}`

Cập nhật thông tin sản phẩm.

**Path Parameters:**
- `id` (int64) - ID của sản phẩm

**Request Body:**
```json
{
  "name": "Updated T-shirt",
  "price": 29.99,
  "category_id": 2
}
```

**Required Fields:**
- `name` (string)
- `price` (float64)
- `category_id` (int64)

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Updated T-shirt",
  "price": 29.99,
  "category_id": 2,
  "created_at": "2026-04-18T09:20:00Z"
}
```

**Error Response:** `500 Internal Server Error`
```
failed to update product
```

**Postman Test:**
```
PUT http://localhost:8081/products/1
Content-Type: application/json

{
  "name": "Premium T-shirt",
  "price": 34.99,
  "category_id": 1
}
```

---

### 6. Delete Product

**DELETE** `/products/{id}`

Xóa sản phẩm.

**Path Parameters:**
- `id` (int64) - ID của sản phẩm

**Response:** `204 No Content`
(No response body)

**Error Response:** `500 Internal Server Error`
```
failed to delete product
```

**Postman Test:**
```
DELETE http://localhost:8081/products/1
```

---

## Postman Collection Quick Reference

### Setup
1. Create new Postman Collection: `Product Service`
2. Set Base URL as variable: `{{PRODUCT_BASE_URL}}` = `http://localhost:8081`

### Requests

```
1. Health Check
   GET {{PRODUCT_BASE_URL}}/health

2. List All Products
   GET {{PRODUCT_BASE_URL}}/products

3. Get Product (ID: 1)
   GET {{PRODUCT_BASE_URL}}/products/1

4. Create Product
   POST {{PRODUCT_BASE_URL}}/products
   Body (JSON):
   {
     "name": "New Product",
     "price": 99.99,
     "category_id": 1
   }

5. Update Product (ID: 1)
   PUT {{PRODUCT_BASE_URL}}/products/1
   Body (JSON):
   {
     "name": "Updated Product",
     "price": 149.99,
     "category_id": 1
   }

6. Delete Product (ID: 1)
   DELETE {{PRODUCT_BASE_URL}}/products/1
```

---

## Data Types

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `id` | int64 | ID sản phẩm (auto-generated) | 1 |
| `name` | string | Tên sản phẩm | "Basic T-shirt" |
| `price` | float64 | Giá sản phẩm | 19.99 |
| `category_id` | int64 | ID danh mục | 1 |
| `created_at` | string (RFC3339) | Thời gian tạo | "2026-04-18T09:20:00Z" |

---

## HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request thành công (GET, PUT) |
| 201 | Created | Tạo resource thành công (POST) |
| 204 | No Content | Request thành công nhưng không có response body (DELETE) |
| 400 | Bad Request | Dữ liệu request không hợp lệ |
| 404 | Not Found | Resource không tồn tại |
| 500 | Internal Server Error | Lỗi server (database, etc.) |

---

## Environment Variables for Postman

Tạo environment trong Postman với các biến:

```
PRODUCT_BASE_URL: http://localhost:8081
PRODUCT_ID: 1
```

---

## Testing Workflow

### Workflow 1: Create → Read → Update → Delete

```
1. GET /health (verify service is running)
2. GET /products (list all)
3. POST /products (create new)
4. GET /products/{new_id} (get detail)
5. PUT /products/{new_id} (update)
6. DELETE /products/{new_id} (delete)
7. GET /products/{new_id} (verify deleted - should be 404)
```

### Workflow 2: Test Error Cases

```
1. POST /products with invalid data (missing name)
   → Expect: 400 Bad Request

2. POST /products with negative price
   → Expect: 400 Bad Request

3. GET /products/99999 (non-existent ID)
   → Expect: 404 Not Found

4. PUT /products with invalid ID format
   → Expect: 400 Bad Request
```

---

## Notes

- **Database Persistence:** Tất cả dữ liệu được lưu trong MySQL database, không phải mock data.
- **ID Auto-increment:** Khi tạo product, ID được tự động sinh và không thể set trong request body.
- **Timestamps:** `created_at` được tự động set khi product được tạo trong database.
- **Price Validation:** Giá phải là số dương (> 0).
- **Category ID:** Hiện tại không validate category ID có tồn tại hay không (FE cần ensure valid category).
- **Timezone:** Tất cả timestamps là UTC (RFC3339 format).
