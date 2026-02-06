"use client";

import { useTranslations } from "next-intl";
import { useOrders } from "@/entities/order/model/use-orders";
import { OrderRow } from "@/entities/order/ui/OrderRow";
import { CancelOrderButton } from "@/features/cancel-order/CancelOrderButton";
import { EditOrderDialog } from "@/features/edit-order/EditOrderDialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

export function OpenOrdersTable() {
  const t = useTranslations("order");
  const { data: orders, isLoading } = useOrders(50, 0);

  const pending = orders?.filter((o) => o.status === "PENDING");

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{t("openOrders")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!pending?.length) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{t("openOrders")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground py-4 text-center text-sm">
            {t("noOpenOrders")}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">
          {t("openOrders")} ({pending.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
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
            {pending.map((o) => (
              <TableRow key={o.id}>
                <OrderRow order={o} />
                <td className="px-3 py-2">
                  <div className="flex gap-1">
                    <EditOrderDialog order={o} />
                    <CancelOrderButton orderId={o.id} />
                  </div>
                </td>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
