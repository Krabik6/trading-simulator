import { api } from "@/shared/api/client";
import type { Price } from "../model/types";

export async function fetchPrices(): Promise<Price[]> {
  const { data } = await api.get<Price[]>("/prices");
  return data;
}
