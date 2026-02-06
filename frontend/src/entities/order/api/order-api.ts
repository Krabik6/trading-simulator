import { api } from "@/shared/api/client";
import type { Order, CreateOrderRequest, UpdateOrderRequest } from "../model/types";

export async function fetchOrders(
  limit = 50,
  offset = 0,
): Promise<Order[]> {
  const { data } = await api.get<Order[]>("/orders", {
    params: { limit, offset },
  });
  return data;
}

export async function fetchOrder(id: number): Promise<Order> {
  const { data } = await api.get<Order>(`/orders/${id}`);
  return data;
}

export async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const { data } = await api.post<Order>("/orders", req);
  return data;
}

export async function updateOrder(
  id: number,
  req: UpdateOrderRequest,
): Promise<Order> {
  const { data } = await api.patch<Order>(`/orders/${id}`, req);
  return data;
}

export async function cancelOrder(
  id: number,
): Promise<{ status: string }> {
  const { data } = await api.delete<{ status: string }>(`/orders/${id}`);
  return data;
}
