import type { WSEnvelope, WSMessageType } from './types';

type MessageHandler = (envelope: WSEnvelope) => void;

class WSClient {
  private socket: WebSocket | null = null;
  private url: string;
  private handlers = new Map<WSMessageType | 'ALL', Set<MessageHandler>>();
  private heartbeatTimer: any = null;
  private reconnectTimer: any = null;
  private retryCount = 0;
  private maxRetries = 10;
  private heartbeatInterval = 10000; // 10s match server's (approx)
  private isClosing = false;

  constructor(url: string) {
    this.url = url;
  }

  public connect() {
    if (this.socket) {
      console.warn('WS already connecting or connected');
      return;
    }

    this.isClosing = false;
    console.log(`Connecting to WS: ${this.url}`);
    
    try {
      this.socket = new WebSocket(this.url);
    } catch (err) {
      console.error('Failed to create WebSocket instance', err);
      this.scheduleReconnect();
      return;
    }

    this.socket.onopen = () => {
      console.log('WS Connection established');
      this.retryCount = 0;
      this.startHeartbeat();
      // Notify all listeners
      this.emit({ type: 'HEARTBEAT', seq: 0, room_id: 'hall', payload: { status: 'OPEN' } });
    };

    this.socket.onmessage = (event) => {
      try {
        const envelope: WSEnvelope = JSON.parse(event.data);
        this.emit(envelope);
      } catch (err) {
        console.error('Failed to parse WS message', err);
      }
    };

    this.socket.onclose = (event) => {
      console.log(`WS Connection closed: ${event.code} ${event.reason}`);
      this.cleanup();
      this.socket = null; // 重要：断开后置空才能重新 connect
      if (!this.isClosing) {
        this.scheduleReconnect();
      }
    };

    this.socket.onerror = (err) => {
      console.error('WS Connection error', err);
      // close will be called automatically
    };
  }

  public disconnect() {
    this.isClosing = true;
    this.cleanup();
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  public send(type: WSMessageType, payload: any = {}, room_id = 'hall') {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.warn(`Cannot send message ${type}, socket not open`);
      return false;
    }

    const envelope: WSEnvelope = {
      type,
      seq: 0, // Server might ignore or handle this
      room_id,
      payload,
    };

    this.socket.send(JSON.stringify(envelope));
    return true;
  }

  public on(type: WSMessageType | 'ALL', handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  private emit(envelope: WSEnvelope) {
    // Specific type handlers
    this.handlers.get(envelope.type)?.forEach(h => h(envelope));
    // Catch-all handlers
    this.handlers.get('ALL')?.forEach(h => h(envelope));
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      this.send('HEARTBEAT');
    }, this.heartbeatInterval);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return;
    if (this.retryCount >= this.maxRetries) {
      console.error('Max WS reconnect retries reached');
      return;
    }

    this.retryCount++;
    const delay = Math.min(1000 * Math.pow(2, this.retryCount), 30000); // Exponential backoff
    console.log(`Scheduling WS reconnect in ${delay}ms (attempt ${this.retryCount})`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private cleanup() {
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  public isConnected() {
    return this.socket?.readyState === WebSocket.OPEN;
  }
}

// Singleton for easier access, but could be put in a context or store
const wsUrl = import.meta.env.VITE_WS_URL || `ws://${window.location.hostname}:8080/ws`;
export const wsClient = new WSClient(wsUrl);
