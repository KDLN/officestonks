// WebSocket client for OfficeStonks real-time updates

// Types
export interface StockUpdate {
  id: number;
  symbol: string;
  current_price: number;
  previous_price: number;
  timestamp: string;
}

// Event listeners for WebSocket events
type MessageListener = (data: StockUpdate) => void;
type ConnectionListener = () => void;
type ErrorListener = (error: Event) => void;

class StockWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private connected: boolean = false;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private reconnectInterval: number = 3000; // 3 seconds
  private messageListeners: MessageListener[] = [];
  private connectListeners: ConnectionListener[] = [];
  private disconnectListeners: ConnectionListener[] = [];
  private errorListeners: ErrorListener[] = [];

  constructor(baseUrl: string = '') {
    // Default to current host if not specified
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = baseUrl || window.location.host;
    this.url = `${wsProtocol}//${host}/ws`;
  }

  // Connect to the WebSocket server
  public connect(): void {
    if (this.ws) {
      this.disconnect();
    }
    
    try {
      this.ws = new WebSocket(this.url);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.connected = true;
        this.connectListeners.forEach(listener => listener());
      };
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as StockUpdate;
          this.messageListeners.forEach(listener => listener(data));
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };
      
      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.connected = false;
        this.disconnectListeners.forEach(listener => listener());
        this.scheduleReconnect();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.errorListeners.forEach(listener => listener(error));
      };
      
    } catch (err) {
      console.error('Failed to connect to WebSocket:', err);
      this.scheduleReconnect();
    }
  }

  // Disconnect from the WebSocket server
  public disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.connected = false;
    }
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // Schedule a reconnection attempt
  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    
    this.reconnectTimer = setTimeout(() => {
      console.log('Attempting to reconnect WebSocket...');
      this.connect();
    }, this.reconnectInterval);
  }

  // Check if the WebSocket is connected
  public isConnected(): boolean {
    return this.connected;
  }

  // Add a listener for stock update messages
  public onMessage(listener: MessageListener): () => void {
    this.messageListeners.push(listener);
    return () => {
      this.messageListeners = this.messageListeners.filter(l => l !== listener);
    };
  }

  // Add a listener for connection events
  public onConnect(listener: ConnectionListener): () => void {
    this.connectListeners.push(listener);
    return () => {
      this.connectListeners = this.connectListeners.filter(l => l !== listener);
    };
  }

  // Add a listener for disconnection events
  public onDisconnect(listener: ConnectionListener): () => void {
    this.disconnectListeners.push(listener);
    return () => {
      this.disconnectListeners = this.disconnectListeners.filter(l => l !== listener);
    };
  }

  // Add a listener for error events
  public onError(listener: ErrorListener): () => void {
    this.errorListeners.push(listener);
    return () => {
      this.errorListeners = this.errorListeners.filter(l => l !== listener);
    };
  }
}

// Create a singleton instance
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";
export const stockWebSocket = new StockWebSocket(API_BASE_URL);