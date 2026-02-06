import { api } from "@/shared/api/client";
import type { Price, Ticker24h } from "../model/types";

export async function fetchPrices(): Promise<Price[]> {
  const { data } = await api.get<Price[]>("/prices");
  return data;
}

export async function fetchTicker24h(): Promise<Ticker24h[]> {
  const { data } = await api.get<Ticker24h[]>("/ticker24h");
  return data;
}
