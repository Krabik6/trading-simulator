"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchPositions,
  closePosition,
  updateTPSL,
} from "../api/position-api";
import type { UpdateTPSLRequest } from "./types";

export const positionKeys = {
  all: ["positions"] as const,
};

export function usePositions() {
  return useQuery({
    queryKey: positionKeys.all,
    queryFn: fetchPositions,
  });
}

export function useClosePosition() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => closePosition(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["positions"] });
      qc.invalidateQueries({ queryKey: ["account"] });
      qc.invalidateQueries({ queryKey: ["trades"] });
    },
  });
}

export function useUpdateTPSL() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...req }: UpdateTPSLRequest & { id: number }) =>
      updateTPSL(id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["positions"] });
    },
  });
}
