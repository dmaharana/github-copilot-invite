# GitHub Copilot Invitation Manager

A REST API service that manages GitHub organization teams and Copilot invitations with Smartsheet license validation.

## Features

- List GitHub organizations
- List teams within an organization
- Create new GitHub teams
- Send GitHub Copilot invitations with license validation
- Smartsheet integration for license tracking

## Setup

1. Copy the configuration template:
```bash
cp config.yaml.template config.yaml
```

2. Edit `config.yaml` with your credentials:
```yaml
github:
  token: "your-github-token-here"  # Token with org admin permissions

smartsheet:
  token: "your-smartsheet-token-here"  # Smartsheet API token
  sheet_id: 123456789  # Your Smartsheet ID containing license information

server:
  port: 8080
  environment: "development"
```

3. Build and run:
```bash
go build
./github-copilot-invite
```

## API Endpoints

All API endpoints (except `/health`) require authentication using a Bearer token. Include the token in the Authorization header:

```
Authorization: Bearer your-api-token-here
```

### List Organizations
```
GET /api/v1/orgs
Authorization: Bearer your-api-token-here
```

### List Teams in Organization
```
GET /api/v1/orgs/{org}/teams
Authorization: Bearer your-api-token-here
```

### Create Team in Organization
```
POST /api/v1/orgs/{org}/teams
Authorization: Bearer your-api-token-here
Content-Type: application/json

{
  "name": "team-name",
  "description": "Team description",
  "privacy": "closed"
}
```

### Send Copilot Invitation
```
POST /api/v1/copilot/invite
Authorization: Bearer your-api-token-here
Content-Type: application/json

{
  "organization": "org-name",
  "team": "team-name",
  "username": "github-username"
}
```

## Smartsheet Configuration

The Smartsheet should have the following columns:
1. Organization Name (Text)
2. Available Licenses (Number)

## Error Handling

The API returns appropriate HTTP status codes:
- 200: Success
- 400: Bad Request (invalid input)
- 401: Unauthorized
- 403: Forbidden
- 409: Conflict (no licenses available)
- 500: Internal Server Error
