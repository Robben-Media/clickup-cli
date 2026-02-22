# Authorization

## Overview

Implement user identity and OAuth token endpoints.

**Why**: `whoami` validates API key configuration. OAuth token exchange is needed for apps that use OAuth flow (not personal tokens).

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /user | Get Authorized User | GetAuthorizedUser | v2 |
| POST | /oauth/token | Get Access Token | GetAccessToken | v2 |

## User Stories

### US-001: Get Current User

**CLI Command:** `clickup auth whoami`

**JSON Output:**
```json
{"user": {"id": 1, "username": "jeremy", "email": "jeremy@example.com", "color": "#4194f6", "profilePicture": "https://..."}}
```

**Plain Output (TSV):** Headers: `ID	USERNAME	EMAIL`
```
1	jeremy	jeremy@example.com
```

**Human-Readable:**
```
Authenticated as:
  ID: 1
  Username: jeremy
  Email: jeremy@example.com
```

**Acceptance Criteria:**
- [ ] Returns the user associated with the current API key
- [ ] Useful for verifying API key is valid

### US-002: OAuth Token Exchange

**CLI Command:** `clickup auth token --client-id <id> --client-secret <secret> --code <auth_code>`

**JSON Output:**
```json
{"access_token": "pk_...", "token_type": "Bearer"}
```

**Plain Output (TSV):** Headers: `ACCESS_TOKEN	TOKEN_TYPE`

**Acceptance Criteria:**
- [ ] Exchanges authorization code for access token
- [ ] Returns bearer token
- [ ] This is for OAuth apps, not personal token users

## Request/Response Types

```go
type AuthorizedUserResponse struct {
    User AuthUser `json:"user"`
}

type AuthUser struct {
    ID             int    `json:"id"`
    Username       string `json:"username"`
    Email          string `json:"email"`
    Color          string `json:"color,omitempty"`
    ProfilePicture string `json:"profilePicture,omitempty"`
}

type OAuthTokenRequest struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    Code         string `json:"code"`
}

type OAuthTokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
}
```

## Edge Cases

- Invalid API key returns 401 — show clear error
- OAuth endpoint doesn't need Authorization header
- Token exchange is a one-time operation per auth code

## Feedback Loops

### Unit Tests
```go
func TestAuthService_Whoami(t *testing.T) { /* get current user */ }
func TestAuthService_Token(t *testing.T)  { /* OAuth exchange */ }
```
