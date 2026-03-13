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

1. **YouTube API Key**: You need a Google Cloud API key with the following APIs enabled:
   - YouTube Data API v3
   - YouTube Analytics API

2. **Go**: Version 1.21 or higher

## Getting Your YouTube Credentials

### Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable billing (required for API access)

### Step 2: Enable Required APIs

1. Navigate to "APIs & Services" > "Library"
2. Search and enable:
   - **YouTube Data API v3**
   - **YouTube Analytics API**

### Step 3: Create OAuth 2.0 credentials

1. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
2. Choose "Desktop app" as application type
3. Download the JSON file and save it as `client_secret.json`
4. Make sure to add some branding details and _Test Users_:
   - Navigate to console.cloud.google.com
   - Select your project
   - In the left sidebar, go to: APIs & Services → OAuth consent screen
   - Scroll down to the "Test users" section
   - Click "+ ADD USERS"
   - Enter your Google/Gmail email address (the one associated with your YouTube channel)
   - Click Save
   - _Make sure "User Type" is set to "External" (unless you have a Google Workspace account)_

## Installation

### macOS

```bash
brew tap thejasmeetsingh/yt-analytics-mcp https://github.com/thejasmeetsingh/yt-analytics-mcp
brew install --cask yt-analytics-mcp  # or: brew install yt-analytics-mcp
```

### Linux

```bash
wget https://github.com/thejasmeetsingh/yt-analytics-mcp/releases/download/v{tag}/yt-analytics-mcp_{tag}_linux_x86_64.deb
sudo dpkg -i yt-analytics-mcp_{tag}_linux_x86_64.deb
```

**Note:**
Replace `{tag}` with the specific version number from the repository (e.g., `0.1.0`).
This tag corresponds to the release version available on the GitHub repository.

### Windows

```bash
scoop bucket add thejasmeetsingh https://github.com/thejasmeetsingh/yt-analytics-mcp
scoop install thejasmeetsingh/yt-analytics-mcp
```

### Generate Auth Token

```bash
yt-analytics-mcp -credentials=/path/to/client_secret.json -token
```

## MCP Configuration

Add this to your MCP settings file (e.g., `claude_desktop_config.json` for Claude Desktop):

```json
{
  "mcpServers": {
    "youtube-analytics": {
      "command": "yt-analytics-mcp", // "yt-analytics-mcp.exe" for windows
      "args": ["-credentials", "/path/to/client_secret.json"]
    }
  }
}
```

For Claude Desktop, the config file is typically located at:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

## Usage

Once the server is running and configured with your MCP client, you can use these tools:

### 1. List Your Channels

First, get your channel IDs:

```
Use the list_channels tool to see all channels
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
Get analytics for channel UC_x5XG1OV2P6uZZ5FSM9Ttw
from 2024-01-01 to 2024-12-31
```

### 5. List Videos

```
List the last 50 videos from channel UC_x5XG1OV2P6uZZ5FSM9Ttw
```

### 6. Get Video Analytics

```
Get analytics for video IDs: dQw4w9WgXcQ,abc123def
from the last 30 days
```

### 7. Force Refresh (Bypass Cache)

```
Get analytics for channel UC_x5XG1OV2P6uZZ5FSM9Ttw
with force_refresh set to true
```

### Comparison Tools Examples 🆕

#### 8. Compare Two Time Periods

```
Compare my channel performance:
December 2024 (2024-12-01 to 2024-12-31)
vs November 2024 (2024-11-01 to 2024-11-30)
```

This will show you:

- Views, watch time, engagement changes
- Subscriber growth comparison
- Percentage changes with trend indicators
- Key insights on what improved or declined

#### 9. Compare Multiple Videos

```
Compare these videos: dQw4w9WgXcQ,abc123def,xyz789ghi
from the last 30 days
```

This shows:

- Side-by-side performance metrics
- Which video got most views, engagement, retention
- Detailed breakdown of each video
- Identifies your top performers

#### 10. Analyze Publishing Schedule

```
Which days work best for my uploads on channel UC_x5XG1OV2P6uZZ5FSM9Ttw?
Look back 90 days
```

This reveals:

- Average views by day of week
- Engagement rates per day
- Recommendations on best/worst days
- Data-driven publishing strategy

#### 11. Compare Video Formats

```
Compare my video formats on channel UC_x5XG1OV2P6uZZ5FSM9Ttw
with keywords: tutorial,review,tips,shorts
from the last 90 days
```

This analyzes:

- Performance by content type
- Which formats get best views/retention
- Engagement by format
- Strategic recommendations on what to create more of

## Available Tools

### Basic Analytics

1. **list_channels** - List all your YouTube channels
2. **get_channel_details** - Get channel information and statistics
3. **get_channel_analytics** - Comprehensive analytics for a date range
4. **get_video_list** - List videos from a channel
5. **get_video_analytics** - Detailed analytics for specific videos

### Comparison Tools 🆕

6. **compare_channel_periods** - Compare performance between two time periods
7. **compare_videos** - Side-by-side comparison of multiple videos
8. **compare_publishing_schedule** - Find which days perform best for uploads
9. **compare_video_formats** - Compare performance by content type (tutorials, reviews, etc.)

## Available Metrics

### Channel Analytics

- Total views
- Total watch time (hours)
- Total likes
- Total comments
- Total shares
- Subscribers gained
- Subscribers lost
- Net subscriber change
- Average view duration
- Average view percentage
- Engagement rate
- Top traffic sources

### Video Analytics

- Views
- Likes
- Comments
- Watch time
- Average view duration
- Average view percentage
- Shares

### Comparison Metrics

- Period-over-period growth
- Video performance rankings
- Day-of-week performance patterns
- Content format performance analysis

## Rate Limiting & Caching

- **Rate Limit**: 5 requests per second - Burst is 10
- **Cache TTL**: 5 minutes
- **Force Refresh**: Use `force_refresh: true` to bypass cache

## Troubleshooting

### "API key not valid" Error

- Verify your API key is correct
- Check that YouTube Data API v3 and YouTube Analytics API are enabled
- Ensure the API key isn't restricted from these APIs

### "Channel not found" Error

- Make sure you're using the correct channel ID (starts with UC)
- Verify the channel belongs to the Google account that created the API key

### "Quota exceeded" Error

- YouTube API has daily quota limits
- Wait for quota to reset (usually at midnight Pacific Time)
- Consider implementing longer cache times or reducing requests

### Analytics Data is Empty

- Make sure your channel has the YouTube Partner Program enabled
- Some analytics require a minimum threshold of views/watch time
- Check that your date range is valid

### Authentication Issues

- The YouTube Analytics API requires OAuth 2.0 for some features
- This implementation uses API Key authentication which has limitations
- For full analytics access, consider implementing OAuth 2.0

## API Quota Information

YouTube Data API v3 has a quota limit of **10,000 units per day** by default:

- `list` operations: 1 unit
- `search` operations: 100 units

The rate limiter in this server helps prevent hitting rate limits, but monitor your daily quota in the Google Cloud Console.

## Contributing

Feel free to open issues or submit pull requests for improvements!

## License

MIT License - feel free to use this for your projects!

---

Once configured with Claude Desktop, you can ask:

- "Show me my YouTube channels"
- "What are my analytics for the last month?"
- "Compare my top 5 videos"
- "How did my channel perform in Q4 2024?"
- **"Compare December vs November performance"** 🆕
- **"Which days should I upload videos?"** 🆕
- **"Do my tutorials or reviews perform better?"** 🆕
- **"Show me my best performing content types"** 🆕

The server will automatically use the appropriate tools and return formatted markdown responses!
