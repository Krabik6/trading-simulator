"use client";

import { env } from "@/shared/config/env";

export type WsMessageType = "prices" | "position" | "position_close" | "pong";

export interface WsMessage<T = unknown> {
  type: WsMessageType;
  data: T;
  timestamp: string;
}

type WsListener = (msg: WsMessage) => void;

export class WebSocketManager {
  private ws: WebSocket | null = null;
  private listeners = new Set<WsListener>();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private token: string | null = null;

  connect(token?: string | null) {
    this.token = token ?? null;
    this.reconnectAttempts = 0;
    this.doConnect();
  }

  private doConnect() {
    this.cleanup();
    const url = this.token ? `${env.wsUrl}?token=${this.token}` : env.wsUrl;
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.pingTimer = setInterval(() => {
        this.ws?.send(JSON.stringify({ type: "ping" }));
      }, 30_000);
    };

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as WsMessage;
        if (msg.type === "pong") return;
        this.listeners.forEach((fn) => fn(msg));
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this.clearPing();
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return;
    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30_000);
    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => this.doConnect(), delay);
  }

  subscribe(listener: WsListener) {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  disconnect() {
    this.cleanup();
    this.ws?.close();
    this.ws = null;
    this.reconnectAttempts = this.maxReconnectAttempts; // prevent reconnect
  }

  private clearPing() {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  private cleanup() {
    this.clearPing();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}

export const wsManager = new WebSocketManager();
