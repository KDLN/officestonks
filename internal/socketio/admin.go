package socketio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// AdminServer provides Socket.IO admin dashboard functionality
type AdminServer struct {
	socketServer *SocketIOServer
	authEnabled  bool
	username     string
	password     string
}

// AdminStats represents the statistics sent to the admin UI
type AdminStats struct {
	ServerID         string                 `json:"serverId"`
	ServerName       string                 `json:"serverName"`
	Version          string                 `json:"version"`
	Uptime          int64                  `json:"uptime"`
	ClientsCount     int                    `json:"clientsCount"`
	NamespacesCount  int                    `json:"namespacesCount"`
	RoomsCount       int                    `json:"roomsCount"`
	Namespaces       []NamespaceInfo        `json:"namespaces"`
	Clients          []ClientStatsInfo      `json:"clients"`
	Events           map[string]interface{} `json:"events"`
}

// NamespaceInfo represents namespace statistics
type NamespaceInfo struct {
	Name         string `json:"name"`
	ClientsCount int    `json:"clientsCount"`
	RoomsCount   int    `json:"roomsCount"`
}

// ClientStatsInfo represents client information for admin UI
type ClientStatsInfo struct {
	ID          string            `json:"id"`
	Connected   int64             `json:"connected"`
	UserID      int               `json:"userId"`
	Username    string            `json:"username"`
	IP          string            `json:"ip"`
	Transport   string            `json:"transport"`
	Rooms       []string          `json:"rooms"`
	Data        map[string]string `json:"data"`
}

// NewAdminServer creates a new admin server for Socket.IO monitoring
func NewAdminServer(socketServer *SocketIOServer) *AdminServer {
	// Get admin credentials from environment
	adminUser := os.Getenv("SOCKETIO_ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	
	adminPass := os.Getenv("SOCKETIO_ADMIN_PASS")
	if adminPass == "" {
		adminPass = "officestonks2024"
		log.Printf("⚠️ Using default admin password. Set SOCKETIO_ADMIN_PASS env var for production")
	}

	admin := &AdminServer{
		socketServer: socketServer,
		authEnabled:  true,
		username:     adminUser,
		password:     adminPass,
	}

	log.Println("🔧 Socket.IO Admin server configured")
	return admin
}

// ServeAdminUI serves the Socket.IO admin interface
func (a *AdminServer) ServeAdminUI(w http.ResponseWriter, r *http.Request) {
	// Basic authentication for admin UI
	if a.authEnabled {
		user, pass, ok := r.BasicAuth()
		if !ok || user != a.username || pass != a.password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Socket.IO Admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Serve the admin dashboard
	a.serveAdminDashboard(w, r)
}

// serveAdminDashboard serves a comprehensive admin interface for Socket.IO debugging
func (a *AdminServer) serveAdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	stats := a.getServerStats()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Socket.IO Admin - Office Stonks</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; background: #f8f9fa; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 2rem; }
        .header h1 { margin: 0; font-size: 2rem; }
        .header p { margin: 0.5rem 0 0 0; opacity: 0.9; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem; }
        .card { background: white; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); overflow: hidden; }
        .card-header { background: #667eea; color: white; padding: 1rem; font-weight: 600; }
        .card-body { padding: 1rem; }
        .stat-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
        .stat-item { text-align: center; padding: 1rem; background: #f8f9fa; border-radius: 8px; }
        .stat-value { font-size: 2rem; font-weight: bold; color: #667eea; }
        .stat-label { font-size: 0.875rem; color: #6c757d; margin-top: 0.25rem; }
        .clients-table { width: 100%%; border-collapse: collapse; margin-top: 1rem; }
        .clients-table th, .clients-table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #e9ecef; }
        .clients-table th { background: #f8f9fa; font-weight: 600; color: #495057; }
        .clients-table tr:hover { background: #f8f9fa; }
        .badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 0.25rem; font-size: 0.75rem; font-weight: 600; }
        .badge-success { background: #d4edda; color: #155724; }
        .badge-info { background: #cce7ff; color: #004085; }
        .btn { background: #667eea; color: white; border: none; padding: 0.5rem 1rem; border-radius: 0.375rem; cursor: pointer; margin: 0.25rem; }
        .btn:hover { background: #5a67d8; }
        .debug-section { background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-top: 1rem; }
        .debug-section h4 { margin: 0 0 1rem 0; color: #495057; }
        .debug-info { font-family: 'Courier New', monospace; background: white; padding: 1rem; border-radius: 4px; overflow: auto; max-height: 300px; }
        .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%%; margin-right: 0.5rem; }
        .status-connected { background: #28a745; }
        .status-disconnected { background: #dc3545; }
        .refresh-notice { background: #fff3cd; color: #856404; padding: 1rem; border-radius: 8px; margin-bottom: 1rem; }
    </style>
    <script>
        function refreshStats() {
            window.location.reload();
        }
        
        function testConnection() {
            fetch('/socket.io/test-connection')
                .then(response => response.json())
                .then(data => {
                    alert('Connection test result: ' + JSON.stringify(data, null, 2));
                })
                .catch(error => {
                    alert('Connection test failed: ' + error.message);
                });
        }
        
        // Auto-refresh every 30 seconds
        setInterval(refreshStats, 30000);
    </script>
</head>
<body>
    <div class="header">
        <h1>🎯 Socket.IO Admin Dashboard</h1>
        <p>Office Stonks Real-time Server Monitoring & Debugging</p>
    </div>
    
    <div class="container">
        <div class="refresh-notice">
            <strong>Auto-refresh:</strong> This page refreshes every 30 seconds. 
            <button class="btn" onclick="refreshStats()">🔄 Refresh Now</button>
            <button class="btn" onclick="testConnection()">🧪 Test Connection</button>
        </div>

        <div class="grid">
            <div class="card">
                <div class="card-header">📊 Server Overview</div>
                <div class="card-body">
                    <div class="stat-grid">
                        <div class="stat-item">
                            <div class="stat-value">%d</div>
                            <div class="stat-label">Connected Clients</div>
                        </div>
                        <div class="stat-item">
                            <div class="stat-value">%d</div>
                            <div class="stat-label">Active Namespaces</div>
                        </div>
                        <div class="stat-item">
                            <div class="stat-value">%s</div>
                            <div class="stat-label">Server Status</div>
                        </div>
                        <div class="stat-item">
                            <div class="stat-value">WebSocket + Polling</div>
                            <div class="stat-label">Transport Mode</div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">🔗 Connection Health</div>
                <div class="card-body">
                    <div class="debug-section">
                        <h4>Connection Debugging Info</h4>
                        <div class="debug-info">
                            <strong>Backend URL:</strong> %s<br>
                            <strong>Socket.IO Endpoint:</strong> /socket.io/<br>
                            <strong>WebSocket URL:</strong> wss://%s/socket.io/<br>
                            <strong>Polling Fallback:</strong> https://%s/socket.io/<br>
                            <strong>CORS Origin:</strong> *<br>
                            <strong>Auth Method:</strong> JWT Token in query params<br>
                            <br>
                            <strong>Common Issues:</strong><br>
                            • Token format: Ensure not [object Promise]<br>
                            • CORS: Check Railway domain matches<br>
                            • Transport: WebSocket → Polling fallback<br>
                            • Port: Same port for HTTP and WebSocket<br>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="card-header">👥 Connected Clients (%d)</div>
            <div class="card-body">
                <table class="clients-table">
                    <thead>
                        <tr>
                            <th>Status</th>
                            <th>Socket ID</th>
                            <th>User</th>
                            <th>IP Address</th>
                            <th>Connected At</th>
                            <th>Rooms</th>
                        </tr>
                    </thead>
                    <tbody>
                        %s
                    </tbody>
                </table>
            </div>
        </div>

        <div class="card">
            <div class="card-header">🔍 Full Server Statistics (JSON)</div>
            <div class="card-body">
                <div class="debug-info">
                    <pre>%s</pre>
                </div>
            </div>
        </div>
        
        <div class="card">
            <div class="card-header">🚀 Socket.IO Server Information</div>
            <div class="card-body">
                <p><strong>Office Stonks Socket.IO Server</strong> - Real-time stock market simulation with Railway deployment compatibility.</p>
                <p>This dashboard helps debug WebSocket connections, monitor client activity, and troubleshoot real-time communication issues.</p>
                <p><strong>Default Login:</strong> admin / officestonks2024</p>
            </div>
        </div>
    </div>
</body>
</html>
`, stats.ClientsCount, stats.NamespacesCount, "🟢 Online", 
   "wss://beta.officestonks.com", "beta.officestonks.com", "beta.officestonks.com",
   stats.ClientsCount, a.generateClientsTableHTML(), string(statsJSON))

	w.Write([]byte(html))
}

// generateClientsTableHTML creates HTML table rows for connected clients
func (a *AdminServer) generateClientsTableHTML() string {
	clients := a.socketServer.GetConnectedClients()
	if len(clients) == 0 {
		return `<tr><td colspan="6" style="text-align: center; font-style: italic; color: #6c757d;">No clients connected</td></tr>`
	}

	html := ""
	for _, client := range clients {
		rooms := ""
		for room := range client.Namespaces {
			if rooms != "" {
				rooms += ", "
			}
			rooms += room
		}
		
		html += fmt.Sprintf(`
			<tr>
				<td><span class="status-indicator status-connected"></span><span class="badge badge-success">Connected</span></td>
				<td><code>%s</code></td>
				<td><strong>%s</strong> (#%d)</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>
		`, client.SocketID, client.Username, client.UserID, client.IPAddress, 
		   client.ConnectedAt.Format("15:04:05"), rooms)
	}
	
	return html
}

// getServerStats compiles comprehensive server statistics for admin UI
func (a *AdminServer) getServerStats() AdminStats {
	clients := a.socketServer.GetConnectedClients()
	
	// Build namespace info
	namespaceCounts := make(map[string]int)
	for _, client := range clients {
		for namespace := range client.Namespaces {
			namespaceCounts[namespace]++
		}
	}

	namespaces := []NamespaceInfo{}
	for ns, count := range namespaceCounts {
		namespaces = append(namespaces, NamespaceInfo{
			Name:         ns,
			ClientsCount: count,
			RoomsCount:   1, // Simplified for now
		})
	}

	// Build client info for admin UI
	clientStats := []ClientStatsInfo{}
	for _, client := range clients {
		rooms := []string{}
		for room := range client.Namespaces {
			rooms = append(rooms, room)
		}

		clientStats = append(clientStats, ClientStatsInfo{
			ID:        client.SocketID,
			Connected: client.ConnectedAt.Unix(),
			UserID:    client.UserID,
			Username:  client.Username,
			IP:        client.IPAddress,
			Transport: "auto", // Socket.IO handles transport detection
			Rooms:     rooms,
			Data: map[string]string{
				"userID": strconv.Itoa(client.UserID),
			},
		})
	}

	return AdminStats{
		ServerID:        "officestonks-socketio-1",
		ServerName:      "Office Stonks Socket.IO Server",
		Version:         "1.0.0",
		Uptime:         time.Now().Unix(),
		ClientsCount:    len(clients),
		NamespacesCount: len(namespaceCounts),
		RoomsCount:      len(namespaceCounts), // Simplified
		Namespaces:      namespaces,
		Clients:         clientStats,
		Events: map[string]interface{}{
			"stock_update":      "Real-time stock price updates",
			"chat_message":      "Chat system messages", 
			"admin_announcement": "Admin broadcast messages",
		},
	}
}

// GetAdminStatsEndpoint returns an HTTP handler for admin stats API
func (a *AdminServer) GetAdminStatsEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Basic auth check
		if a.authEnabled {
			user, pass, ok := r.BasicAuth()
			if !ok || user != a.username || pass != a.password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Socket.IO Admin API"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		stats := a.getServerStats()
		json.NewEncoder(w).Encode(stats)
	}
}