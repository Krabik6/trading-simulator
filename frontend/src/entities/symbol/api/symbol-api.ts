import { api } from "@/shared/api/client";
import type { SymbolSpec } from "../model/types";

export async function fetchSymbols(): Promise<SymbolSpec[]> {
  const { data } = await api.get<SymbolSpec[]>("/symbols");
  return data;
}
