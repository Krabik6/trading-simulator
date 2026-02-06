import { api } from "@/shared/api/client";
import type { Account } from "../model/types";

export async function fetchAccount(): Promise<Account> {
  const { data } = await api.get<Account>("/account");
  return data;
}
