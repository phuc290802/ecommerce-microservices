# Ecommerce Microservices

Đây là một demo microservices đơn giản với:
- API Gateway bằng Go
- Product service bằng Go (MySQL optional)
- Order service bằng Go
- Frontend VueJS với Vite
- Docker Compose để chạy toàn bộ hệ thống

## Chạy nhanh

Từ thư mục gốc:

```bash
docker compose up --build
```

Sau khi chạy xong:
- Frontend: http://localhost:5173
- API Gateway: http://localhost:8080
- Product service: http://localhost:8081
- Order service: http://localhost:8082

## API Gateway

Gateway proxy các đường dẫn:
- `/api/products` -> Product service
- `/api/orders` -> Order service

Các tính năng đã triển khai:
- Validate JWT (chỉ kiểm tra hạn, không gọi DB)
- Rate limiting theo IP và theo `user_id` nếu có
- Log request: method, path, status, duration, client_ip
- Retry tối đa 3 lần khi downstream trả lỗi
- Circuit breaker nếu lỗi >50% trong 10s
- Forward header `X-Request-ID` cho downstream

## Tạo JWT test

Bạn có thể tạo token mẫu bằng Python:

```bash
python3 - <<'PY'
import jwt, time
payload = {
    'sub': 'user-123',
    'exp': int(time.time()) + 3600
}
print(jwt.encode(payload, 'supersecret', algorithm='HS256'))
PY
```

Sau đó dùng header:

```
Authorization: Bearer <token>
```

## Lưu ý

Product service có thể kết nối tới MySQL nếu chạy cùng Docker Compose. MySQL sẽ được khởi tạo tại `mysql:3306` và database `ecommerce`.

Nếu bạn muốn thay đổi cấu hình gateway, sửa file `docker-compose.yml` hoặc biến môi trường trong `gateway/Dockerfile`.
