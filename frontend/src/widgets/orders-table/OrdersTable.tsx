"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useOrders } from "@/entities/order/model/use-orders";
import { OrderRow } from "@/entities/order/ui/OrderRow";
import { CancelOrderButton } from "@/features/cancel-order/CancelOrderButton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

export function OrdersTable() {
  const t = useTranslations("order");
  const [tab, setTab] = useState<"all" | "pending">("all");
  const [offset, setOffset] = useState(0);
  const limit = 20;
  const { data: orders, isLoading } = useOrders(limit, offset);

  const filtered =
    tab === "pending"
      ? orders?.filter((o) => o.status === "PENDING")
      : orders;

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between pb-3">
        <CardTitle className="text-base">{t("orders")}</CardTitle>
        <Tabs value={tab} onValueChange={(v) => setTab(v as "all" | "pending")}>
          <TabsList>
            <TabsTrigger value="all">{t("all")}</TabsTrigger>
            <TabsTrigger value="pending">{t("pending")}</TabsTrigger>
          </TabsList>
        </Tabs>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : !filtered?.length ? (
          <p className="text-muted-foreground py-8 text-center text-sm">
            {t("noOrders")}
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
                  <TableHead>{t("leverage")}</TableHead>
                  <TableHead>{t("status")}</TableHead>
                  <TableHead>{t("time")}</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((o) => (
                  <TableRow key={o.id}>
                    <OrderRow order={o} />
                    <td className="px-3 py-2">
                      {o.status === "PENDING" && (
                        <CancelOrderButton orderId={o.id} />
                      )}
                    </td>
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
                disabled={(orders?.length ?? 0) < limit}
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
