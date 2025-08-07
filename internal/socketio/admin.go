package socketio

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
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
	adminFS      http.Handler
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

// AdminUIFiles - we'll embed the admin UI files here
//go:embed admin/dist/*
var adminUIFiles embed.FS

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

	// Create admin file server from embedded files
	adminFS, err := fs.Sub(adminUIFiles, "admin/dist")
	if err != nil {
		log.Printf("⚠️ Admin UI files not found, admin dashboard will not be available")
		adminFS = nil
	}

	admin := &AdminServer{
		socketServer: socketServer,
		authEnabled:  true,
		username:     adminUser,
		password:     adminPass,
	}

	if adminFS != nil {
		admin.adminFS = http.FileServer(http.FS(adminFS))
	}

	// Setup admin namespace event handlers
	admin.setupAdminHandlers()

	return admin
}

// setupAdminHandlers configures admin-specific Socket.IO events
func (a *AdminServer) setupAdminHandlers() {
	// Setup admin namespace (simplified for now)
	// Note: The doquangtan/socket.io library has limited namespace support
	// For now, we'll handle admin functionality through the main namespace with special event prefixes

	log.Println("🔧 Socket.IO Admin handlers configured")
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

	// Serve admin UI files
	if a.adminFS != nil {
		a.adminFS.ServeHTTP(w, r)
	} else {
		// Fallback: serve a basic admin page
		a.serveFallbackAdminPage(w, r)
	}
}

// serveFallbackAdminPage serves a basic admin interface when UI files aren't available
func (a *AdminServer) serveFallbackAdminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	stats := a.getServerStats()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Socket.IO Admin - Office Stonks</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; margin-bottom: 20px; }
        .stats { background: #ecf0f1; padding: 15px; border-radius: 4px; }
        .stat-item { margin: 10px 0; }
        .stat-label { font-weight: bold; color: #2980b9; }
        .clients-table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        .clients-table th, .clients-table td { padding: 8px 12px; text-align: left; border: 1px solid #bdc3c7; }
        .clients-table th { background: #3498db; color: white; }
        .refresh-btn { background: #27ae60; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #229954; }
        pre { background: #2c3e50; color: #ecf0f1; padding: 15px; border-radius: 4px; overflow: auto; }
    </style>
    <script>
        function refreshStats() {
            window.location.reload();
        }
        
        // Auto-refresh every 30 seconds
        setInterval(refreshStats, 30000);
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Socket.IO Admin Dashboard</h1>
            <p>Office Stonks Real-time Server Monitoring</p>
        </div>
        
        <button class="refresh-btn" onclick="refreshStats()">🔄 Refresh Stats</button>
        
        <div class="stats">
            <div class="stat-item">
                <span class="stat-label">Connected Clients:</span> %d
            </div>
            <div class="stat-item">
                <span class="stat-label">Active Namespaces:</span> %d
            </div>
            <div class="stat-item">
                <span class="stat-label">Server Uptime:</span> %s
            </div>
            <div class="stat-item">
                <span class="stat-label">Transport Mode:</span> WebSocket + Polling Fallback
            </div>
        </div>

        <h3>📊 Connected Clients</h3>
        <table class="clients-table">
            <thead>
                <tr>
                    <th>Socket ID</th>
                    <th>User ID</th>
                    <th>Username</th>
                    <th>IP Address</th>
                    <th>Connected At</th>
                    <th>Rooms</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>

        <h3>🔍 Server Statistics (JSON)</h3>
        <pre>%s</pre>
        
        <div style="margin-top: 30px; padding: 15px; background: #d4edda; border-radius: 4px; border: 1px solid #c3e6cb;">
            <strong>🎮 Office Stonks Socket.IO Server</strong><br>
            Real-time stock market simulation with Railway deployment compatibility.<br>
            Admin Dashboard - Monitor connections, broadcast messages, and track performance.
        </div>
    </div>
</body>
</html>
`, stats.ClientsCount, stats.NamespacesCount, time.Now().Format("2006-01-02 15:04:05"), a.generateClientsTableHTML(), string(statsJSON))

	w.Write([]byte(html))
}

// generateClientsTableHTML creates HTML table rows for connected clients
func (a *AdminServer) generateClientsTableHTML() string {
	clients := a.socketServer.GetConnectedClients()
	if len(clients) == 0 {
		return `<tr><td colspan="6" style="text-align: center; font-style: italic; color: #7f8c8d;">No clients connected</td></tr>`
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
				<td>%s</td>
				<td>%d</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>
		`, client.SocketID, client.UserID, client.Username, client.IPAddress, 
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