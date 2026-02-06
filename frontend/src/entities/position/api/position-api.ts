import { api } from "@/shared/api/client";
import type { Position, UpdateTPSLRequest } from "../model/types";

export async function fetchPositions(): Promise<Position[]> {
  const { data } = await api.get<Position[]>("/positions");
  return data;
}

export async function fetchPosition(id: number): Promise<Position> {
  const { data } = await api.get<Position>(`/positions/${id}`);
  return data;
}

export async function closePosition(
  id: number,
  quantity?: string,
): Promise<{ status: string; realized_pnl: string; closed_quantity: string }> {
  const body = quantity ? { quantity } : undefined;
  const { data } = await api.post<{
    status: string;
    realized_pnl: string;
    closed_quantity: string;
  }>(`/positions/${id}/close`, body);
  return data;
}

export async function updateTPSL(
  id: number,
  req: UpdateTPSLRequest,
): Promise<Position> {
  const { data } = await api.patch<Position>(`/positions/${id}`, req);
  return data;
}
