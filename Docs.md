# ChefShare API Documentation

This document provides comprehensive details about the ChefShare API endpoints, specifically focused on the API changes related to the recent improvements, including the new email verification flow.

## Table of Contents

1. [Authentication Routes](#authentication-routes)
   - [Register User](#register-user)
   - [Login User](#login-user)
   - [Refresh Token](#refresh-token)
   - [Get Authenticated User](#get-authenticated-user)
   - [Logout User](#logout-user)
   - [Password Reset Flow](#password-reset-flow)
   - [Email Verification Flow](#email-verification-flow)
2. [User Profile Routes](#user-profile-routes)
   - [Update User](#update-user)
   - [Update Password](#update-password)
3. [Health Check](#health-check)
4. [Response Codes](#response-codes)
5. [Rate Limiting](#rate-limiting)

---

## Authentication Routes

### Register User

Register a new user in the system.

- **URL:** `POST /api/v1/auth/register`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "username": "newuser",
    "email": "user@example.com",
    "password": "Password123!",
    "bio": "A short bio about the user.",
      "first_name": "John",
      "last_name": "Doe",
      "profile_picture": "https://example.com/profile.jpg",
      "email_verified": false,
      "created_at": "..."
    }
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Invalid request body or validation failure.
    ```json
    {
      "error": "detailed validation error"
    }
    ```
  - `409 Conflict`: Username or email already exists.
    ```json
    {
      "error": "username already exists"
    }
    ```
    OR
    ```json
    {
      "error": "email already exists"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Login User

Authenticates a user and provides access and refresh tokens.

- **URL:** `POST /api/v1/auth/login`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "Password123!"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "login successful",
    "tokens": {
      "access_token": "...",
      "refresh_token": "..."
    },
    "user": {
      "user_id": "...",
      "username": "newuser",
      "email": "user@example.com",
      "bio": "A short bio about the user.",
      "first_name": "John",
      "last_name": "Doe",
      "profile_picture": "https://example.com/profile.jpg",
      "email_verified": false,
      "created_at": "...",
      "last_login": "..."
    }
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Invalid request body.
    ```json
    {
      "error": "invalid request format"
    }
    ```
  - `401 Unauthorized`: Invalid email or password.
    ```json
    {
      "error": "invalid email or password"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Refresh Token

Refreshes an expired access token using a valid refresh token.

- **URL:** `POST /api/v1/auth/token/refresh`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "refresh_token": "..."
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "token refreshed",
    "tokens": {
      "access_token": "...",
      "refresh_token": "..."  // Note: The refresh token is rotated for security
    }
  }
  ```

- **Error Responses:**
  - `401 Unauthorized`: Invalid or expired refresh token.
    ```json
    {
      "error": "invalid or expired refresh token"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Get Authenticated User

Retrieves the profile of the currently authenticated user.

- **URL:** `GET /api/v1/auth/me`
- **Authentication:** Bearer Token in `Authorization` header
- **Request Body:** None

- **Success Response (200 OK):**
  ```json
  {
    "user": {
      "user_id": "...",
      "username": "newuser",
      "email": "user@example.com",
      "bio": "A short bio about the user.",
      "first_name": "John",
      "last_name": "Doe",
      "profile_picture": "https://example.com/profile.jpg",
      "email_verified": true,
      "created_at": "..."
    }
  }
  ```

- **Error Responses:**
  - `401 Unauthorized`: Missing, invalid, or expired access token.
    ```json
    {
      "error": "authentication required"
    }
    ```
    OR
    ```json
    {
      "error": "invalid or expired token"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Logout User

Logs out the user by revoking their refresh token.

- **URL:** `POST /api/v1/auth/logout`
- **Authentication:** Bearer Token in `Authorization` header
- **Request Body:**
  ```json
  {
    "refresh_token": "..."
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "logout successful"
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Missing refresh token.
    ```json
    {
      "error": "refresh token is required"
    }
    ```
  - `401 Unauthorized`: Missing, invalid, or expired access token.
    ```json
    {
      "error": "authentication required"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Email Verification Flow

#### Verify Email

Verifies a user's email address using a verification token.

- **URL:** `POST /api/v1/auth/verify-email/confirm`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "token": "verification-token-from-email"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "email verified successfully"
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Invalid token format.
    ```json
    {
      "error": "token is required"
    }
    ```
  - `404 Not Found`: Token doesn't exist.
    ```json
    {
      "error": "invalid or expired verification token"
    }
    ```
  - `410 Gone`: Token has expired.
    ```json
    {
      "error": "verification link has expired, please request a new one"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

#### Resend Verification Email (Unauthenticated)

Resends the verification email to a user's email address.

- **URL:** `POST /api/v1/auth/verify-email/resend`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "if your email is registered and not verified, a verification email will be sent"
  }
  ```

  Note: For security reasons, this success response is always returned whether or not the email exists or is already verified.

- **Error Responses:**
  - `400 Bad Request`: Invalid email format.
    ```json
    {
      "error": "invalid email format"
    }
    ```
  - `429 Too Many Requests`: Rate limit exceeded.
    ```json
    {
      "message": "too many verification attempts, please try again later"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

#### Request Verification Email (Authenticated)

Requests a new verification email for the currently authenticated user.

- **URL:** `POST /api/v1/auth/verify-email/request`
- **Authentication:** Bearer Token in `Authorization` header
- **Request Body:** None

- **Success Response (200 OK):**
  ```json
  {
    "message": "verification email sent"
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Email is already verified.
    ```json
    {
      "error": "email is already verified"
    }
    ```
  - `401 Unauthorized`: Missing, invalid, or expired access token.
    ```json
    {
      "error": "unauthorized"
    }
    ```
  - `429 Too Many Requests`: Rate limit exceeded.
    ```json
    {
      "message": "too many verification attempts, please try again later"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Password Reset Flow

#### Request Password Reset

Initiates the password reset process by sending an OTP to the user's email.

- **URL:** `POST /api/v1/auth/password/reset/request`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "if your email is registered, we've sent a password reset code"
  }
  ```
  
  Note: To prevent email enumeration, this response is always returned, regardless of whether the email exists.

- **Error Responses:**
  - `400 Bad Request`: Invalid email format.
    ```json
    {
      "error": "invalid email format"
    }
    ```
  - `429 Too Many Requests`: Rate limit exceeded.
    ```json
    {
      "message": "too many password reset attempts, please try again later"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

#### Verify OTP and Reset Password

Verifies the OTP and resets the user's password.

- **URL:** `POST /api/v1/auth/password/reset/confirm`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "otp": "123456",
    "password": "NewPassword456!"
  }
  ```
  - `otp` (string, required): 6-digit code sent to the user's email.
  - `password` (string, required): The new password. Must meet the same requirements as during registration.

- **Success Response (200 OK):**
  ```json
  {
    "message": "password reset successful",
    "sessions_revoked": true,
    "info": "please log in with your new password"
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Invalid request body, invalid OTP, or password does not meet requirements.
    ```json
    {
      "error": "invalid OTP format"
    }
    ```
    OR
    ```json
    {
      "error": "password must be at least 8 characters with a number and symbol"
    }
    ```
    OR
    ```json
    {
      "error": "invalid or expired OTP"
    }
    ```
  - `404 Not Found`: User not found.
    ```json
    {
      "error": "user not found"
    }
    ```
  - `429 Too Many Requests`: Rate limit exceeded.
    ```json
    {
      "error": "too many password reset attempts, please try again later"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

#### Resend OTP

Resends the OTP to the user's email.

- **URL:** `POST /api/v1/auth/password/reset/resend`
- **Authentication:** None required
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "if your email is registered, we've sent a new password reset code"
  }
  ```
  Note: To prevent email enumeration, this response is always returned, regardless of whether the email exists.

- **Error Responses:**
  - `400 Bad Request`: Invalid email format.
    ```json
    {
      "error": "invalid email format"
    }
    ```
  - `429 Too Many Requests`: Rate limit exceeded.
    ```json
    {
      "message": "too many password reset attempts, please try again later"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

## User Profile Routes

### Update User

Updates the profile of the authenticated user.

- **URL:** `PUT /api/v1/users/me`
- **Authentication:** Bearer Token in `Authorization` header
- **Request Body:**
  ```json
  {
    "current_password": "Password123!",
    "username": "new_username",
    "first_name": "Johnny",
    "last_name": "Doer",
    "bio": "An updated bio.",
    "profile_picture": "https://example.com/new_profile.jpg"
  }
  ```
  - `current_password` (string, required): For verification.
  - All other fields are optional.

- **Success Response (200 OK):**
  ```json
  {
    "message": "profile updated successfully",
    "user": {
      "user_id": "...",
      "username": "new_username",
      "email": "user@example.com",
      "bio": "An updated bio.",
      "first_name": "Johnny",
      "last_name": "Doer",
      "profile_picture": "https://example.com/new_profile.jpg",
      "created_at": "...",
      "updated_at": "..."
    }
  }
  ```

- **Error Responses:**
  - `400 Bad Request`: Invalid request body or validation failure.
    ```json
    {
      "error": "detailed validation error"
    }
    ```
  - `401 Unauthorized`: Invalid `current_password`.
    ```json
    {
      "error": "invalid password"
    }
    ```
  - `404 Not Found`: User not found.
    ```json
    {
      "error": "user not found"
    }
    ```
  - `409 Conflict`: New username is already taken.
    ```json
    {
      "error": "username already taken"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

### Update Password

Updates the password of the authenticated user.

- **URL:** `PUT /api/v1/users/me/password`
- **Authentication:** Bearer Token in `Authorization` header
- **Request Body:**
  ```json
  {
    "current_password": "Password123!",
    "password": "NewPassword456!"
  }
  ```

- **Success Response (200 OK):**
  ```json
  {
    "message": "password updated successfully"
  }
  ```
  Note: After a successful password change, all existing sessions (refresh tokens) are invalidated.

- **Error Responses:**
  - `400 Bad Request`: Invalid request body or new password does not meet requirements.
    ```json
    {
      "error": "password must be at least 8 characters with at least one number and special character"
    }
    ```
  - `401 Unauthorized`: Invalid `current_password`.
    ```json
    {
      "error": "invalid current password"
    }
    ```
  - `404 Not Found`: User not found.
    ```json
    {
      "error": "user not found"
    }
    ```
  - `500 Internal Server Error`: An unexpected error occurred.
    ```json
    {
      "error": "internal server error"
    }
    ```

## Health Check

Provides a health check endpoint for monitoring the API's status.

- **URL:** `GET /api/v1/health`
- **Authentication:** None required
- **Request Body:** None

- **Success Response (200 OK):**
  ```json
  {
    "status": "ok",
    "timestamp": "2023-09-05T12:34:56Z",
    "dependencies": {
      "database": {
        "status": "ok",
        "message": ""
      }
    }
  }
  ```

- **Error Responses:**
  - If the database is not available, the response will still be 200 OK but will indicate the issue:
    ```json
    {
      "status": "ok",
      "timestamp": "2023-09-05T12:34:56Z",
      "dependencies": {
        "database": {
          "status": "error",
          "message": "connection refused"
        }
      }
    }
    ```

## Response Codes

| Code | Description |
|------|-------------|
| 200 | OK - The request was successful. |
| 201 | Created - The resource was successfully created. |
| 400 | Bad Request - The request could not be understood or was missing required parameters. |
| 401 | Unauthorized - Authentication failed or user doesn't have permissions. |
| 404 | Not Found - Resource was not found. |
| 409 | Conflict - Request conflicts with the current state of the server (e.g., duplicate resources). |
| 429 | Too Many Requests - Rate limit has been exceeded. |
| 500 | Internal Server Error - An error occurred on the server. |

## Rate Limiting

The password reset endpoints are rate-limited to prevent abuse:

1. IP-based rate limiting: 5 requests per IP address per 15 minutes.
2. Email-based rate limiting: 3 requests per email address per hour.

When a rate limit is exceeded, the API will respond with a 429 Too Many Requests status code:

```json
{
  "message": "too many password reset attempts, please try again later"
}
```

It's recommended to implement appropriate UI feedback when this occurs, such as a countdown timer or a message advising the user to try again later.