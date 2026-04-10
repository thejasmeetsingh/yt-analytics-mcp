# YouTube Analytics MCP Server

A Model Context Protocol (MCP) server that provides comprehensive YouTube channel and video analytics using the official YouTube Data API v3 and YouTube Analytics API.

## Features

- 📊 **Channel Analytics**: Views, watch time, subscribers, engagement rates
- 🎥 **Video Analytics**: Individual video performance metrics
- 📈 **Traffic Sources**: Understand where your views come from
- ⚡ **Rate Limiting**: Built-in protection for API quota management
- 💾 **Smart Caching**: 5-minute in-memory cache with force-refresh option
- 📝 **Markdown Output**: LLM-friendly formatted responses

## Prerequisites

**YouTube API Access**: You need a Google Cloud OAuth 2.0 client ID with the following APIs enabled:
  - YouTube Data API v3
  - YouTube Analytics API

## Getting Your YouTube Credentials

### Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable billing (required for API access) — no payment is required for testing

### Step 2: Enable Required APIs

1. Navigate to **APIs & Services** > **Library**
2. Search and enable:
   - YouTube Data API v3
   - YouTube Analytics API

### Step 3: Create OAuth 2.0 Credentials

1. Go to **Credentials** → **Create Credentials** → **OAuth client ID**
2. Choose **Desktop app** as the application type
3. Download the JSON file and save it as `client_secret.json`
4. Add branding details and test users:
   - Navigate to [console.cloud.google.com](https://console.cloud.google.com)
   - Select your project
   - Go to **APIs & Services** → **OAuth consent screen**
   - Scroll down to **Test users** and click **+ ADD USERS**
   - Enter the Gmail address associated with your YouTube channel
   - Click **Save**
   - Ensure **User Type** is set to **External** (unless you have a Google Workspace account)

## Installation

### macOS

```bash
brew tap thejasmeetsingh/yt-analytics-mcp https://github.com/thejasmeetsingh/yt-analytics-mcp
brew install --cask yt-analytics-mcp
```

### Linux

```bash
wget https://github.com/thejasmeetsingh/yt-analytics-mcp/releases/download/v{tag}/yt-analytics-mcp_{tag}_linux_x86_64.deb
sudo dpkg -i yt-analytics-mcp_{tag}_linux_x86_64.deb
```

> Replace `{tag}` with the specific version number from the [releases page](https://github.com/thejasmeetsingh/yt-analytics-mcp/releases) (e.g. `0.1.0`).

### Windows

```bash
scoop bucket add thejasmeetsingh https://github.com/thejasmeetsingh/yt-analytics-mcp
scoop install thejasmeetsingh/yt-analytics-mcp
```

### Configuration

The application requires Google OAuth 2.0 credentials to be provided via environment variables:

#### Required Environment Variables

```bash
export GOOGLE_CLIENT_ID="your_client_id"
export GOOGLE_CLIENT_SECRET="your_client_secret"
export GOOGLE_REDIRECT_URL="http://localhost:8080/callback"  # Optional, defaults to http://localhost:8080/callback
```

Retrieve your `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` from the OAuth 2.0 credentials JSON file you downloaded from Google Cloud Console.

### Generate or Refresh Auth Token

Run this command after setting the environment variables to authenticate with your Google account. You can also re-run it at any time to regenerate an expired or revoked token:

```bash
yt-analytics-mcp -token
```

This command starts an interactive authentication flow. Here's exactly what to expect:

**Step 1 — A URL appears in your terminal.**
After running the command, you'll see output similar to this:

```
=== YouTube API Authorization Required ===

Please visit this URL in your browser:
https://accounts.google.com/o/oauth2/auth?client_id=...

After authorizing, paste the authorization code here:
```

**Step 2 — Open the URL in your browser.**
Copy the full URL and paste it into any browser. Sign in with the Google account that owns your YouTube channel.

**Step 3 — Grant permissions.**
Google will ask you to allow the app access to your YouTube data. Click **Allow**. If you see a warning that the app isn't verified, click **Advanced** → **Go to [app name] (unsafe)** — this is expected for apps in testing mode.

**Step 4 — Copy the code from the redirect page.**
After granting access, Google redirects you to a page that may look broken or blank. That's normal. Look at the URL in your browser's address bar — it will contain a `code=` parameter, for example:

```
http://localhost/?code=4/0AX4XfWh...&scope=...
```

Copy everything after `code=` up to the `&` (or to the end of the URL if there is no `&`). This is your verification code.

**Step 5 — Paste the code into your terminal.**
Go back to your terminal, paste the code at the `After authorizing, paste the authorization code here:` prompt, and press Enter. The app will exchange the code for credentials and save them locally. You won't need to do this again unless your token expires or is revoked.

## MCP Configuration

Add the following to your MCP settings file. For Claude Desktop this is typically `claude_desktop_config.json`:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**macOS / Linux:**

```json
{
  "mcpServers": {
    "youtube-analytics": {
      "command": "yt-analytics-mcp",
      "env": {
        "GOOGLE_CLIENT_ID": "your_client_id_here",
        "GOOGLE_CLIENT_SECRET": "your_client_secret_here",
        "GOOGLE_REDIRECT_URL": "http://localhost:8080/callback"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "youtube-analytics": {
      "command": "yt-analytics-mcp.exe",
      "env": {
        "GOOGLE_CLIENT_ID": "your_client_id_here",
        "GOOGLE_CLIENT_SECRET": "your_client_secret_here",
        "GOOGLE_REDIRECT_URL": "http://localhost:8080/callback"
      }
    }
  }
}
```

## Usage

Once the server is running and configured with your MCP client, you can use the following tools.

### 1. List Your Channels

Get your channel IDs — you'll need these for most other tools:

```
Use the list_channels tool to see all my channels
```

### 2. Get Channel Details

```
Get details for channel ID: UC_x5XG1OV2P6uZZ5FSM9Ttw
```

### 3. Get Channel Analytics (Default: Last 30 Days)

```
Get analytics for channel UC_x5XG1OV2P6uZZ5FSM9Ttw
```

### 4. Get Custom Date Range Analytics

```
Get analytics for channel UC_x5XG1OV2P6uZZ5FSM9Ttw from 2024-01-01 to 2024-12-31
```

### 5. List Videos

```
List the last 50 videos from channel UC_x5XG1OV2P6uZZ5FSM9Ttw
```

### 6. Get Video Analytics

```
Get analytics for video IDs: dQw4w9WgXcQ,abc123def from the last 30 days
```

### 7. Force Refresh (Bypass Cache)

```
Get analytics for channel UC_x5XG1OV2P6uZZ5FSM9Ttw with force_refresh set to true
```

### 8. Compare Two Time Periods

```
Compare my channel performance:
December 2024 (2024-12-01 to 2024-12-31)
vs November 2024 (2024-11-01 to 2024-11-30)
```

Returns:
- Views, watch time, and engagement changes
- Subscriber growth comparison
- Percentage changes with trend indicators
- Key insights on what improved or declined

### 9. Compare Multiple Videos

```
Compare these videos: dQw4w9WgXcQ,abc123def,xyz789ghi from the last 30 days
```

Returns:
- Side-by-side performance metrics
- Which video had the most views, engagement, and retention
- Detailed per-video breakdown
- Top performer highlights

### 10. Analyze Publishing Schedule

```
Which days work best for uploads on channel UC_x5XG1OV2P6uZZ5FSM9Ttw? Look back 90 days
```

Returns:
- Average views by day of week
- Engagement rates per day
- Best and worst days to publish

### 11. Compare Video Formats

```
Compare video formats on channel UC_x5XG1OV2P6uZZ5FSM9Ttw
with keywords: tutorial,review,tips,shorts over the last 90 days
```

Returns:
- Performance by content type
- Which formats get the best views and retention
- Engagement by format
- Strategic recommendations on what to create more of

## Available Tools

### Basic Analytics

| Tool | Description |
|------|-------------|
| `list_channels` | List all your YouTube channels |
| `get_channel_details` | Get channel information and statistics |
| `get_channel_analytics` | Comprehensive analytics for a date range |
| `get_video_list` | List videos from a channel |
| `get_video_analytics` | Detailed analytics for specific videos |

### Comparison Tools

| Tool | Description |
|------|-------------|
| `compare_channel_periods` | Compare performance between two time periods |
| `compare_videos` | Side-by-side comparison of multiple videos |
| `compare_publishing_schedule` | Find which days of the week perform best for uploads |
| `compare_video_formats` | Compare performance by content type (tutorials, reviews, etc.) |

## Available Metrics

### Channel Analytics
- Total views, watch time (hours), likes, comments, shares
- Subscribers gained, lost, and net change
- Average view duration and view percentage
- Engagement rate

### Video Analytics
- Views, likes, comments, shares
- Watch time, average view duration, average view percentage

### Comparison Metrics
- Period-over-period growth rates
- Video performance rankings
- Day-of-week performance patterns
- Content format performance analysis

## Rate Limiting & Caching

| Setting | Value |
|---------|-------|
| Rate limit | 5 requests/second (burst: 10) |
| Cache TTL | 5 minutes |
| Force refresh | Pass `force_refresh: true` to bypass cache |

## Troubleshooting

### "Missing required environment variables" error

- Ensure `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are set in your environment
- Verify the values are correct and not truncated
- Check that the OAuth client has not been deleted or disabled in Google Cloud Console

### "Channel not found" error

- Channel IDs start with `UC` — double-check you're not using a handle or custom URL
- Confirm the channel belongs to the Google account used during `yt-analytics-mcp -token`

### "Quota exceeded" error

- YouTube Data API v3 has a default limit of 10,000 units per day
- Quota resets at midnight Pacific Time
- Monitor your usage in the [Google Cloud Console](https://console.cloud.google.com/) under **APIs & Services** → **Quotas**
- Increase cache TTL or reduce request frequency to stay within limits

### Analytics data is empty

- Some analytics require the YouTube Partner Program to be enabled on your channel
- Certain metrics have a minimum view/watch-time threshold before data appears
- Verify your date range falls within a period that has actual activity

### Authentication errors

- Re-run the token command and follow the step-by-step flow described in the [Generate or Refresh Auth Token](#generate-or-refresh-auth-token) section above — this works for both first-time setup and expired or revoked tokens:
  ```bash
  yt-analytics-mcp -token
  ```
- Make sure your Google account is listed as a test user in the OAuth consent screen
- If you haven't used the app in a while, your token may have expired — regenerating it via the command above is all that's needed

## API Quota Information

YouTube Data API v3 defaults to **10,000 units per day**:

| Operation | Cost |
|-----------|------|
| `list` | 1 unit |
| `search` | 100 units |

The built-in rate limiter helps avoid hitting limits during normal use, but check your daily quota in the Google Cloud Console if you run many requests.

## Contributing

Feel free to open issues or submit pull requests for improvements!

## License

MIT License — feel free to use this in your own projects.

---

Once configured with Claude Desktop, you can ask things like:

- "Show me my YouTube channels"
- "What are my analytics for the last month?"
- "Compare my top 5 videos"
- "How did my channel perform in Q4 2024?"
- "Compare December vs November performance"
- "Which days should I upload videos?"
- "Do my tutorials or reviews perform better?"
- "Show me my best performing content types"