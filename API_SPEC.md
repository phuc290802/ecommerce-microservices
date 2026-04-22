# API Specification (Tổng hợp cho Frontend)

Tài liệu ngắn gọn mô tả các API có sẵn từ các microservice trong repo để đội FE xây giao diện và tích hợp.

---

## Tổng quan luồng xác thực (Auth)
- Đăng ký: `POST /register` — body JSON: `{ "username","email","password","phone" }` → 201/200 JSON `{id, username, email}` hoặc lỗi.
- Đăng nhập: `POST /login` — body JSON: `{ "email","password" }` → 200 JSON `{ "access_token", "expires_in", "username" }` và server đặt cookie `refresh_token` (HttpOnly).
- Làm mới token: `POST /refresh` — đọc `refresh_token` cookie → 200 JSON `{ access_token, expires_in }` và cookie `refresh_token` mới.
- Logout: `POST /logout` — xóa cookie `refresh_token` → 204 No Content.
- Xác thực token nội bộ: `POST /internal/verify` — body `{ "token" }` → trả về claims (JSON) hoặc 401.
- OTP: `POST /otp/request` và `POST /otp/validate` — body có `email`/`phone`, `purpose`, `code`.
- Quên/mật khẩu: `POST /forgot-password` (body `{email}`) và `POST /reset-password` (body `{token,password}`) → 204 No Content.

Notes: access_token là JWT (Bearer) dùng cho API cần xác thực; refresh_token lưu trong cookie HttpOnly.

---

## Category Service (category-service) - `:8085`
- `GET /health` → `{ "status": "ok" }`.
- `GET /categories` → list categories từ database (JSON array of categories).
  - Category fields: `{ id, name, slug, parent_id, created_at }`.
- `POST /categories` → tạo category mới. Body JSON: `{ "name", "slug", "parent_id" }` → 201 JSON created category.
- `GET /categories/{id}` → chi tiết category từ database hoặc 404.
- `PUT /categories/{id}` → cập nhật category. Body JSON: `{ "name", "slug", "parent_id" }` → 200 JSON updated category.
- `DELETE /categories/{id}` → xóa category → 204 No Content.
- `GET /categories/tree` → trả về cây categories (nested structure).
- `POST /categories/rebuild` → rebuild cây categories (internal).
- `GET /categories/slug/{slug}` → tìm category theo slug.

Usage examples:
- List: `GET /categories`
- Tree: `GET /categories/tree`
- Detail: `GET /categories/1`
- Create: `POST /categories` with body `{ "name": "Electronics", "slug": "electronics" }`
- Update: `PUT /categories/1` with body `{ "name": "Home Electronics", "slug": "home-electronics" }`
- Delete: `DELETE /categories/1`

---

## Product Service (product-service) - `:8081`
- `GET /health` → `{ "status": "ok" }`.
- `GET /products` → list sản phẩm từ database (JSON array of products).
  - Product fields: `{ id, name, price, category_id, created_at }`.
- `POST /products` → tạo sản phẩm mới. Body JSON: `{ "name", "price", "category_id" }` → 201 JSON created product.
- `GET /products/{id}` → chi tiết product từ database hoặc 404.
- `PUT /products/{id}` → cập nhật product. Body JSON: `{ "name", "price", "category_id" }` → 200 JSON updated product.
- `DELETE /products/{id}` → xóa product → 204 No Content.

Usage examples:
- List: `GET /products`
- Detail: `GET /products/1`
- Create: `POST /products` with body `{ "name": "T-shirt", "price": 19.99, "category_id": 1 }`
- Update: `PUT /products/1` with body `{ "name": "Premium T-shirt", "price": 24.99, "category_id": 1 }`
- Delete: `DELETE /products/1`

---

## Review Service (review-service) - `:8086`
- `GET /health` → `{ "status": "ok" }`.
- `GET /reviews?product_id={id}` → trả về mảng reviews cho product.
  - Review fields: `{ id, product_id, author, rating, comment, created_at }`.

---

## Order Service (order-service) - `:8082`
- `GET /health` → `{ "status": "ok" }`.
- `GET /orders` → danh sách orders (sample data).
- `GET /orders/{id}` → chi tiết order hoặc 404.
  - Order fields: `{ id, product, quantity, total_cost }`.

---

## Stock Service (stock-service) - `:8087`
- `GET /health` → `{ "status": "ok" }`.
- `GET /stock?product_id={id}` → trả về stock info hoặc 404.
  - Stock fields: `{ product_id, available, quantity }`.

---

## Admin Service (admin-service) - `:8088`
- `POST /login` — body `{ email, password }` → JSON `{ token, user }` (admin user), dùng cho UI admin.
- Protected endpoints (require admin headers `X-Admin-ID` và `X-Admin-Role` từ Gateway):
  - `GET /users` → list administrators (JSON array).
  - `POST /users/create` — body JSON admin `{ username, email, password, role }` → 201 Created.
  - `GET /audit-logs` → list audit logs.
  - `GET /dashboard/stats` → dashboard stats JSON.

Notes: hiện implementation dùng header-based auth (gateway-trusted). FE admin cần gọi qua Gateway hoặc set header tương ứng khi phát triển.

---

## BFF (backend-for-frontend) - `:8083`
- `GET /health` → `{ "status": "ok" }`.
- `GET /summary?product_id={id}` → Aggregated JSON kết hợp dữ liệu từ product, category, review, stock.
  - Summary structure (tóm tắt):
    - `item`: `{ id, title, unit_price, category_id, created_date }`
    - `category`: `{ id, name, info }`
    - `reviews`: array of `{ reviewer, stars, text, date }`
    - `stock_status`: `{ available, quantity, status_message }`
    - `aggregated_at`: ISO date string
- `POST /graphql` — body `{ "query": "..." }` → GraphQL result (useful cho UI phức tạp).

Flow example (FE wants product page):
1. FE requests `GET /summary?product_id=123` on BFF.
2. BFF calls `GET /products/123`, `GET /categories/{catId}`, `GET /reviews?product_id=123`, `GET /stock?product_id=123`, aggregates và trả về JSON.

---

## Category Service (category-service) - `:8085`
Chức năng chính:
- CRUD danh mục (dùng cho Admin UI): `name`, `slug`, `parent_id`, `description`.
- Lấy cây danh mục (category tree) — cache Redis, TTL 1 giờ.
- Lấy sản phẩm theo `category_id` (gọi sang product service và lọc kết quả).
- Lấy category theo `slug` (SEO friendly).
- Rebuild category tree (xóa cache và trả về cây mới).

Endpoints:
- `GET /health` → `{ "status": "ok" }`.
- `GET /categories` → danh sách categories.
- `POST /categories` → tạo mới category. Body JSON: `{ "name", "slug", "parent_id", "description" }` → 201 JSON created category.
- `GET /categories/{id}` → chi tiết category.
- `PUT /categories/{id}` → cập nhật category (body giống POST) → 204 No Content.
- `DELETE /categories/{id}` → xóa category → 204 No Content.
- `GET /categories/slug/{slug}` → lấy category bằng slug.
- `GET /categories/tree` → trả về cây danh mục (cached TTL 1 giờ).
- `POST /categories/rebuild` → rebuild tree và cập nhật cache.
- `GET /categories/{id}/products` → trả về danh sách sản phẩm thuộc category (gọi sang `product-service`).

Ghi chú cho FE:
- Admin UI sẽ dùng endpoints CRUD để quản lý category (hoặc gọi qua Admin service / Gateway nếu cần xác thực khác).
- Dùng `GET /categories/tree` để hiển thị menu/cây danh mục; cache ở server giúp giảm tải.


## HTTP status / error behaviors
- 200: success with JSON body.
- 201 / 204: created / no content for some write endpoints.
- 400: bad request (missing param or invalid body).
- 401: unauthorized (invalid credentials or token), 403 forbidden for admin insufficient role.
- 404: resource not found.
- 5xx: upstream/internal errors.

---

## Cấu trúc chung request/response (mẫu)
- JSON request example (login):

```json
{ "email": "user@example.com", "password": "secret" }
```

- JSON response example (product detail):

```json
{
  "id": 1,
  "name": "Basic T-shirt",
  "price": 19.99,
  "category_id": 1,
  "created_at": "2026-04-18T09:20:00Z"
}
```

---

## Ghi chú triển khai cho FE
- Sử dụng `BFF /summary` cho trang product list/detail để lấy tất cả dữ liệu cần thiết trong 1 request.
- Dùng cookie `refresh_token` (HttpOnly) và header `Authorization: Bearer <access_token>` cho các request cần auth.
- Admin UI: gửi qua Gateway (hoặc mô phỏng header `X-Admin-ID`/`X-Admin-Role` khi phát triển cục bộ).
- Nếu cần thêm chi tiết (ví dụ các payload create/update cụ thể), tôi có thể mở rộng file này với ví dụ body đầy đủ.

---

## Hành động tiếp theo (tùy chọn)
- Thêm ví dụ curl cho mỗi endpoint
- Sinh OpenAPI / Swagger dựa trên nội dung hiện tại
- Mở rộng chi tiết các schema request/response

Hãy cho biết yêu cầu mở rộng bạn cần.