"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SymbolSelector } from "@/features/symbol-selector/SymbolSelector";
import { PlaceOrderForm } from "@/features/place-order/PlaceOrderForm";
import { useAccount } from "@/entities/account/model/use-account";
import { formatUSD } from "@/shared/lib/format";

export function OrderPanel() {
  const t = useTranslations("order");
  const { data: account } = useAccount();

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t("placeOrder")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <SymbolSelector />
        {account && (
          <div className="text-muted-foreground text-xs">
            {t("availableMargin")}: {formatUSD(account.available_margin)} USDT
          </div>
        )}
        <PlaceOrderForm />
      </CardContent>
    </Card>
  );
}
