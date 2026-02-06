"use client";

import { useTranslations } from "next-intl";
import { useAccount } from "@/entities/account/model/use-account";
import { BalanceDisplay } from "@/entities/account/ui/BalanceDisplay";
import { PnLDisplay } from "@/entities/position/ui/PnLDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { formatPercent } from "@/shared/lib/format";

export function AccountSummary() {
  const t = useTranslations("account");
  const { data: account, isLoading } = useAccount();

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{t("accountSummary")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-8 w-full" />
          ))}
        </CardContent>
      </Card>
    );
  }

  if (!account) return null;

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t("accountSummary")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <BalanceDisplay balance={account.balance} label={t("balance")} />
        <BalanceDisplay balance={account.equity} label={t("equity")} />
        <Separator />
        <BalanceDisplay balance={account.used_margin} label={t("usedMargin")} />
        <BalanceDisplay
          balance={account.available_margin}
          label={t("availableMargin")}
        />
        <Separator />
        <div className="flex flex-col gap-0.5">
          <span className="text-muted-foreground text-xs">
            {t("unrealizedPnl")}
          </span>
          <PnLDisplay
            value={account.unrealized_pnl}
            percent={
              parseFloat(account.balance) > 0
                ? (parseFloat(account.unrealized_pnl) /
                    parseFloat(account.balance)) *
                  100
                : 0
            }
          />
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-muted-foreground text-xs">
            {t("marginRatio")}
          </span>
          <span className="font-mono text-sm">
            {formatPercent(account.margin_ratio)}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
