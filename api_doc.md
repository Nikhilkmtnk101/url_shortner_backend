# URL Shortener API Documentation

This document provides details of the endpoints available in the **URL Shortener API** collection.

## Base URL
```
{{base_url}}
```

## Endpoints

### 1. Signup
**POST** `/auth/signup`

**Request Body:**
```json
{
  "name": "Nikhil Kumar",
  "email": "nikhil.kumar.civ17@itbhu.ac.in",
  "password": "password"
}
```

**Description:** Registers a new user.

---

### 2. Verify Registration OTP
**POST** `/auth/verify-registration-otp`

**Request Body:**
```json
{
  "email": "nikhil.kumar.civ17@itbhu.ac.in",
  "otp": "162297"
}
```

**Description:** Verifies the OTP for registration.

---

### 3. Login
**POST** `/auth/login`

**Request Body:**
```json
{
  "email": "nikhil.kumar.civ17@itbhu.ac.in",
  "password": "password"
}
```

**Description:** Authenticates a user and returns a token.

---

### 4. Forgot Password
**POST** `/auth/login`

**Request Body:**
```json
{
  "email": "nikhil.kumar.civ17@itbhu.ac.in"
}
```

**Description:** Initiates the password reset process.

---

### 5. Refresh Token
**POST** `/auth/refresh-token`
**From Cookie it takes jwt refresh token**

**Description:** Refreshes the authentication token.

---

### 6. Reset Password
**POST** `/auth/reset-password`

**Request Body:**
```json
{
  "email": "nikhilkmtnk21@gmail.com",
  "new_password": "password"
}
```

**Description:** Resets the user password.

---

### 7. Logout
**POST** `/auth/logout`

**Request Body:**
```json
{
  "email": "nikhilkmtnk21@gmail.com",
  "password": "password"
}
```

**Description:** Logs out a user.

---

### 8. Generate Short URL
**POST** `/url`

**Headers:**
- Authorization: Bearer `YOUR_JWT_TOKEN`
- Content-Type: application/json

**Request Body:**
```json
{
  "long_url": "https://jwt.io/",
  "expires_days": 30,
  "password": "password",
  "alias": "xyzwsk"
}
```

**Description:** Generates a short URL with an optional expiration.

---

### 9. Generate Bulk Short URLs
**POST** `/url/bulk`

**Headers:**
- Authorization: Bearer `YOUR_JWT_TOKEN`
- Content-Type: application/json

**Request Body:**
```json
[
  {
    "long_url": "https://jwt.io/",
    "expires_days": 30,
    "password": "password",
    "alias": "xyzwsk2"
  },
  {
    "long_url": "https://jwt.io/",
    "expires_days": 30,
    "password": "password",
    "alias": "xyzwskq"
  }
]
```

**Description:** Generates multiple short URLs in bulk.

---

### 10. Analytics API
**GET** `/url`

**Headers:**
- Authorization: Bearer `YOUR_JWT_TOKEN`
- Content-Type: application/json

**Description:** Provides analytics for URLs.

---

### 11. Redirect URL
**GET** `/url/{shortCode}`

**Headers:**
- Content-Type: application/json

**Description:** Redirects to the original long URL associated with the short code.

---

### 12. Generate QR Code for Short URL
**GET** `/url/qr/{shortCode}`

**Headers:**
- Content-Type: application/json

**Description:** Generates a QR code for the given short URL.

**Response:**
Returns an image in `image/png` format.

---

## Example Usage

### Generate Short URL (cURL)
```bash
curl -X POST http://localhost:8080/urls \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "long_url": "https://example.com/very/long/url",
    "expires_days": 30
  }'
```

