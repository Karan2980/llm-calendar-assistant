# 🗓️ LLM Calendar Assistant

A Go-based intelligent calendar assistant that uses Large Language Models to help you create and manage daily schedules. The application integrates with Google Calendar and GitHub Models API to provide AI-powered scheduling capabilities.

## ✨ Features

- 🤖 **AI-Powered Scheduling**: Uses GitHub Models API to generate intelligent daily schedules
- 📅 **Google Calendar Integration**: Seamlessly syncs with your Google Calendar
- 🔐 **OAuth2 Authentication**: Secure authentication with Google services
- 🌐 **Web API**: RESTful API for easy integration
- 📱 **JSON Response Format**: Structured task responses with ISO 8601 timestamps
- 🔄 **Token Management**: Automatic token refresh and management

## 🏗️ Project Structure

```
llm-calendar-assistant/
├── cmd/
│   └── api/
│       └── main.go              # Main application entry point
├── internal/
│   ├── ai/
│   │   └── github.go            # GitHub Models API integration
│   ├── auth/
│   │   └── env_auth.go          # Environment-based authentication
│   └── models/
│       └── [model files]        # Data models
├── .env                         # Environment configuration (create this)
├── go.mod                       # Go module dependencies
├── go.sum                       # Go module checksums
└── README.md                    # This file
```

## 🚀 Getting Started

### Prerequisites

- Go 1.19 or higher
- Google Cloud Platform account
- GitHub account with access to GitHub Models
- Google Calendar API enabled

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Karan2980/llm-calendar-assistant.git
   cd llm-calendar-assistant
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the project root:
   ```env
   # Google OAuth2 Configuration
   GOOGLE_CLIENT_ID=your_google_client_id_here
   GOOGLE_CLIENT_SECRET=your_google_client_secret_here
   GOOGLE_REDIRECT_URL=http://localhost:8080/callback

   # Google OAuth2 Tokens (will be populated after authentication)
   GOOGLE_ACCESS_TOKEN=
   GOOGLE_REFRESH_TOKEN=
   GOOGLE_TOKEN_EXPIRY=

   # GitHub Configuration
   GITHUB_TOKEN=your_github_personal_access_token_here

   # Server Configuration
   PORT=8080
   ```

### 🔧 Configuration Setup

#### Google Calendar API Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Calendar API
4. Go to "Credentials" and create OAuth 2.0 Client IDs
5. Add `http://localhost:8080/callback` as an authorized redirect URI
6. Copy the Client ID and Client Secret to your `.env` file

#### GitHub Models API Setup

1. Go to [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Generate a new token with appropriate scopes
3. Add the token to your `.env` file as `GITHUB_TOKEN`

### 🏃‍♂️ Running the Application

1. **Start the server**
   ```bash
   go run cmd/api/main.go
   ```

2. **Complete OAuth authentication**
   - The application will guide you through the Google OAuth flow
   - Save the provided tokens to your `.env` file when prompted

3. **Access the API**
   - Server runs on `http://localhost:8080` (or your configured PORT)

## 📚 API Usage

### Generate Daily Schedule

The AI assistant creates daily schedules based on your prompts and integrates them with Google Calendar.

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/schedule \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a productive work schedule for tomorrow with meetings, coding time, and breaks",
    "model": "gpt-4"
  }'
```

**Example Response:**
```json
[
  {
    "summary": "Morning standup meeting",
    "start": "2024-01-15T09:00:00Z",
    "end": "2024-01-15T09:30:00Z"
  },
  {
    "summary": "Deep work - coding session",
    "start": "2024-01-15T10:00:00Z",
    "end": "2024-01-15T12:00:00Z"
  },
  {
    "summary": "Lunch break",
    "start": "2024-01-15T12:00:00Z",
    "end": "2024-01-15T13:00:00Z"
  }
]
```

## 🔐 Authentication Flow

The application uses OAuth2 for Google Calendar access:

1. **Initial Setup**: Run the application for the first time
2. **Browser Authentication**: Complete Google OAuth in your browser
3. **Token Storage**: Save the provided tokens to your `.env` file
4. **Automatic Refresh**: The app handles token refresh automatically

## 🛠️ Development

### Building

```bash
go build -o llm-calendar-assistant cmd/api/main.go
```

### Running Tests

```bash
go test ./...
```

### Code Structure

- **`cmd/api/main.go`**: Application entry point and server setup
- **`internal/ai/github.go`**: GitHub Models API integration for AI responses
- **`internal/auth/env_auth.go`**: Environment-based authentication handling
- **`internal/models/`**: Data models and structures

## 🔧 Configuration Options

| Environment Variable | Description | Required |
|---------------------|-------------|----------|
| `GOOGLE_CLIENT_ID` | Google OAuth2 Client ID | Yes |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 Client Secret | Yes |
| `GOOGLE_REDIRECT_URL` | OAuth2 redirect URL | Yes |
| `GOOGLE_ACCESS_TOKEN` | Google access token | Auto-generated |
| `GOOGLE_REFRESH_TOKEN` | Google refresh token | Auto-generated |
| `GOOGLE_TOKEN_EXPIRY` | Token expiry timestamp | Auto-generated |
| `GITHUB_TOKEN` | GitHub Personal Access Token | Yes |
| `PORT` | Server port | No (default: 8080) |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request


## 🆘 Troubleshooting

### Common Issues

1. **"No access token configured"**
   - Run the application and complete the OAuth flow
   - Save the provided tokens to your `.env` file

2. **"GitHub Models API returned status 401"**
   - Check your `GITHUB_TOKEN` in the `.env` file
   - Ensure the token has the required permissions

3. **"No .env file found"**
   - Create a `.env` file in the project root
   - The app looks for `.env` in current, parent, or grandparent directories

### Getting Help

- Check the console output for detailed error messages
- Ensure all environment variables are properly set
- Verify your Google Cloud and GitHub API credentials

## 🙏 Acknowledgments

- Google Calendar API for calendar integration
- GitHub Models for AI capabilities
- Go OAuth2 library for authentication
- All contributors and users of this project
