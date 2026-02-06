"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useTrades } from "@/entities/trade/model/use-trades";
import { TradeRow } from "@/entities/trade/ui/TradeRow";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

export function TradesTable() {
  const t = useTranslations("trade");
  const [offset, setOffset] = useState(0);
  const limit = 20;
  const { data: trades, isLoading } = useTrades(limit, offset);

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t("tradeHistory")}</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : !trades?.length ? (
          <p className="text-muted-foreground py-8 text-center text-sm">
            {t("noTrades")}
          </p>
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("symbol")}</TableHead>
                  <TableHead>{t("side")}</TableHead>
                  <TableHead>{t("type")}</TableHead>
                  <TableHead>{t("qty")}</TableHead>
                  <TableHead>{t("price")}</TableHead>
                  <TableHead>{t("pnl")}</TableHead>
                  <TableHead>{t("fee")}</TableHead>
                  <TableHead>{t("time")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {trades.map((tr) => (
                  <TableRow key={tr.id}>
                    <TradeRow trade={tr} />
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <div className="mt-4 flex justify-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={offset === 0}
                onClick={() => setOffset((o) => Math.max(0, o - limit))}
              >
                {t("prev")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={(trades?.length ?? 0) < limit}
                onClick={() => setOffset((o) => o + limit)}
              >
                {t("next")}
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
