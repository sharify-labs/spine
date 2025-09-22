# Spine
The web panel for the Sharify platform.

Spine works alongside other Sharify components:
- **[Canvas](https://github.com/sharify-labs/canvas)**: Serves uploaded content to end users
- **[Zephyr](https://github.com/sharify-labs/zephyr)**: Core upload API and storage service
- **[Sharify-Go](https://github.com/sharify-labs/sharify-go)**: Go SDK for programmatic uploads
- **[Sharify-Desktop](https://github.com/sharify-labs/sharify-desktop)**: Cross-platform desktop app

_Note: This project is not actively maintained and should not be used in a production environment._

## What it does
Spine handles the web interface side of Sharify - user authentication, domain management, token generation, and proxying requests to [Zephyr](https://github.com/sharify-labs/zephyr).<br>
When you log into the Sharify web panel and create custom domains or upload content through the browser, you're interacting with Spine.<br>

## Key Features
- Discord OAuth2 authentication with session management
- Custom domain/subdomain registration for users
- API token generation and reset functionality
- ShareX configuration file generation
- Basic web dashboard via HTMX for dynamic UI updates without full page reloads
- Proxying upload requests to [Zephyr](https://github.com/sharify-labs/zephyr) with JWT authentication

## API endpoints

All endpoints require Discord authentication via session cookies:

#### Authentication
```bash
# Discord OAuth2 flow
GET  /auth/discord           # Redirect to Discord
GET  /auth/discord/callback  # Handle OAuth2 callback
```

#### User Management
```bash
# Web interface
GET  /dashboard              # Main user dashboard
GET  /api/v1/reset-token     # Generate new API token for Zephyr
GET  /api/v1/config/:type    # Download ShareX config (files/pastes/redirects)
```

#### Domain Management
```bash
# List available root domains
GET  /api/v1/domains

# Manage user's custom domains
GET     /api/v1/hosts        # List user's domains
POST    /api/v1/hosts        # Create new subdomain
DELETE  /api/v1/hosts/:name  # Delete domain
```

#### Zephyr Proxy Routes
```bash
# These forward directly to Zephyr with user's JWT
GET     /api/v1/uploads      # List uploads
POST    /api/v1/uploads      # Create upload
DELETE  /api/v1/uploads      # Delete uploads
```

### Authentication Types

| Type               | Purpose              | Format                                        |
|--------------------|----------------------|-----------------------------------------------|
| **Session Cookie** | Web panel access     | Encrypted session with Discord user data      |
| **JWT Token**      | Web-to-Zephyr auth   | Short-lived JWT signed with ECDSA private key |
| **API Token**      | Direct Zephyr access | `sfy_<id>_<key>` format for external tools    |

### ShareX Integration

Generates `.sxcu` configuration files that include:
- An API token for Zephyr authentication
- Available domains as dropdown options
- Prompt fields for custom secrets and expiration times

## Development Setup

1. Copy environment configuration:
```bash
cp .env.example .env
```

2. Generate required keys:
```bash
make keys  # Generates JWT keys, session keys, and admin keys
```

3. Configure Discord OAuth2:
    - [Create a Discord application](https://discord.com/developers/applications)
    - Set `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, and callback URL

4. Set up the database and other services:
    - Set up a [Turso](https://docs.turso.tech/introduction) database
    - Configure `TURSO_DSN`
    - Set `ZEPHYR_ADMIN_KEY` to match Zephyr's `ADMIN_KEY_HASH`
    - Configure `SENTRY_DSN` for error tracking

5. Run the server with `make run` or `air` (for hot reloads)