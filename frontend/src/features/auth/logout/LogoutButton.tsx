"use client";

import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useAuthStore } from "@/shared/lib/auth-store";
import { usePositionStore } from "@/entities/position/model/position-store";
import { usePriceStore } from "@/entities/price/model/price-store";
import { getQueryClient } from "@/shared/api/query-client";
import { Button } from "@/components/ui/button";
import { LogOut } from "lucide-react";

export function LogoutButton() {
  const t = useTranslations("auth");
  const router = useRouter();
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = () => {
    logout();
    usePositionStore.getState().clear();
    usePriceStore.setState({ prices: {} });
    getQueryClient().clear();
    router.push("/login");
  };

  return (
    <Button variant="ghost" size="sm" onClick={handleLogout}>
      <LogOut className="mr-1 h-4 w-4" />
      {t("logout")}
    </Button>
  );
}
