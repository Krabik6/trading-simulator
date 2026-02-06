"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchOrders, createOrder, cancelOrder } from "../api/order-api";
import type { CreateOrderRequest } from "./types";

export const orderKeys = {
  all: ["orders"] as const,
  list: (limit: number, offset: number) =>
    ["orders", "list", limit, offset] as const,
};

export function useOrders(limit = 50, offset = 0) {
  return useQuery({
    queryKey: orderKeys.list(limit, offset),
    queryFn: () => fetchOrders(limit, offset),
  });
}

export function useCreateOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateOrderRequest) => createOrder(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["positions"] });
      qc.invalidateQueries({ queryKey: ["account"] });
    },
  });
}

export function useCancelOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => cancelOrder(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["account"] });
    },
  });
}
