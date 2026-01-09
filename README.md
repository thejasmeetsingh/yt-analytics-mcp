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

## Getting Your YouTube API Key

### Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable billing (required for API access)

### Step 2: Enable Required APIs

1. Navigate to "APIs & Services" > "Library"
2. Search and enable:
   - **YouTube Data API v3**
   - **YouTube Analytics API**

### Step 3: Create API Key

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "API Key"
3. Copy your API key
4. (Optional but recommended) Restrict the API key to only YouTube APIs

### Step 4: Link YouTube Channel (Important!)

For the Analytics API to work, you need to link your YouTube channel to the Google Cloud Project:

1. Go to your [YouTube Studio](https://studio.youtube.com)
2. Click Settings > Channel > Advanced settings
3. Under "Google Account", make sure the same Google account used for the API key has access

**Note**: The API key will only work with channels owned by or managed by the Google account that created the API key.

## Installation

### 1. Clone/Create the Project

```bash
mkdir youtube-mcp-server
cd youtube-mcp-server
```

### 2. Initialize Go Module

Create `go.mod`:

```go
module youtube-mcp-server

go 1.21

require (
	google.golang.org/api v0.154.0
)
```

### 3. Download Dependencies

```bash
go mod download
```

### 4. Set Up Environment Variable

**Linux/Mac:**

```bash
export YOUTUBE_API_KEY="your-api-key-here"
```

**Windows (PowerShell):**

```powershell
$env:YOUTUBE_API_KEY="your-api-key-here"
```

**Or create a `.env` file:**

```bash
YOUTUBE_API_KEY=your-api-key-here
```

### 5. Build the Server

```bash
go build -o youtube-mcp-server
```

## MCP Configuration

Add this to your MCP settings file (e.g., `claude_desktop_config.json` for Claude Desktop):

```json
{
  "mcpServers": {
    "youtube-analytics": {
      "command": "/path/to/youtube-mcp-server",
      "env": {
        "YOUTUBE_API_KEY": "your-api-key-here"
      }
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

- **Rate Limit**: 10 requests per second
- **Cache TTL**: 5 minutes
- **Force Refresh**: Use `force_refresh: true` to bypass cache

## Project Structure

```
youtube-mcp-server/
├── main.go              # Main server implementation
├── tools.go             # MCP tools definitions
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── README.md            # This file
```

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

## Future Enhancements

Potential improvements for future versions:

- [ ] OAuth 2.0 authentication for full analytics access
- [ ] Demographics data support
- [ ] Revenue analytics (requires OAuth)
- [ ] Comparison tools (channel vs channel, video vs video)
- [ ] Export to CSV/JSON
- [ ] Persistent caching (Redis/File-based)
- [ ] Pagination for large video lists
- [ ] Real-time subscriber count
- [ ] Thumbnail and metadata management

## Contributing

Feel free to open issues or submit pull requests for improvements!

## License

MIT License - feel free to use this for your projects!

---

## Quick Start Example

```bash
# Set your API key
export YOUTUBE_API_KEY="AIza..."

# Build
go build -o youtube-mcp-server

# The server communicates via stdin/stdout (MCP protocol)
# It's meant to be used with an MCP client like Claude Desktop
```

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
