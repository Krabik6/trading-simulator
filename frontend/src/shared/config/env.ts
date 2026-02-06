export const env = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081",
  wsUrl: process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8081/ws",
} as const;
