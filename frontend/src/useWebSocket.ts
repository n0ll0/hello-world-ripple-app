// WebSocket connection utility - one connection per event type
const WS_BASE = (import.meta.env?.VITE_WS_BASE_URL as string | undefined) ?? "ws://localhost:8080";

export class WebSocketClient {
	private ws: WebSocket | null = null;
	private reconnectTimer: number | null = null;
	private reconnectDelay = 3000;
	private url: string;
	private messageHandlers: Set<(message: any) => void> = new Set();
	private connectionHandlers: Set<() => void> = new Set();
	private disconnectionHandlers: Set<() => void> = new Set();
	private eventType: string;

	constructor(path: string) {
		this.url = new URL(path, WS_BASE).toString();
		this.eventType = path.split('/').pop() || 'unknown';
	}

	connect() {
		if (this.ws?.readyState === WebSocket.OPEN) {
			return;
		}

		this.ws = new WebSocket(this.url);

		this.ws.onopen = () => {
			console.log(`[${this.eventType}] WebSocket connected`);
			this.connectionHandlers.forEach((handler) => handler());
		};

		this.ws.onmessage = (event) => {
			try {
				const message = JSON.parse(event.data);
				this.messageHandlers.forEach((handler) => handler(message));
			} catch (err) {
				// If not JSON, pass as is
				this.messageHandlers.forEach((handler) => handler(event.data));
			}
		};

		this.ws.onerror = (error) => {
			console.error(`[${this.eventType}] WebSocket error:`, error);
		};

		this.ws.onclose = () => {
			console.log(`[${this.eventType}] WebSocket disconnected`);
			this.disconnectionHandlers.forEach((handler) => handler());
			this.scheduleReconnect();
		};
	}

	disconnect() {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
	}

	send(data: any) {
		if (this.ws?.readyState === WebSocket.OPEN) {
			const message = typeof data === "string" ? data : JSON.stringify(data);
			this.ws.send(message);
		} else {
			console.warn("WebSocket is not connected. Cannot send message.");
		}
	}

	onMessage(handler: (message: any) => void) {
		this.messageHandlers.add(handler);
		return () => this.messageHandlers.delete(handler);
	}

	onConnect(handler: () => void) {
		this.connectionHandlers.add(handler);
		return () => this.connectionHandlers.delete(handler);
	}

	onDisconnect(handler: () => void) {
		this.disconnectionHandlers.add(handler);
		return () => this.disconnectionHandlers.delete(handler);
	}

	private scheduleReconnect() {
		if (this.reconnectTimer) {
			return;
		}
		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null;
			console.log(`[${this.eventType}] Attempting to reconnect WebSocket...`);
			this.connect();
		}, this.reconnectDelay) as unknown as number;
	}

	isConnected(): boolean {
		return this.ws?.readyState === WebSocket.OPEN;
	}
}

// Export singleton instances for each event type
export const todoCreatedWS = new WebSocketClient("/ws/todos/created");
export const todoUpdatedWS = new WebSocketClient("/ws/todos/updated");
export const todoDeletedWS = new WebSocketClient("/ws/todos/deleted");
