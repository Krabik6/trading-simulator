import { api } from "@/shared/api/client";
import type { Trade } from "../model/types";

export async function fetchTrades(limit = 50, offset = 0): Promise<Trade[]> {
  const { data } = await api.get<Trade[]>("/trades", {
    params: { limit, offset },
  });
  return data;
}
