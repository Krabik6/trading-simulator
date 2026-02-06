"use client";

import { useTranslations } from "next-intl";
import { usePositions } from "@/entities/position/model/use-positions";
import { PositionRow } from "@/entities/position/ui/PositionRow";
import { ClosePositionButton } from "@/features/close-position/ClosePositionButton";
import { UpdateTPSLDialog } from "@/features/update-tpsl/UpdateTPSLDialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

export function PositionsTable() {
  const t = useTranslations("position");
  const { data: positions, isLoading } = usePositions();

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t("openPositions")}</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : !positions?.length ? (
          <p className="text-muted-foreground py-8 text-center text-sm">
            {t("noPositions")}
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("symbol")}</TableHead>
                <TableHead>{t("side")}</TableHead>
                <TableHead>{t("qty")}</TableHead>
                <TableHead>{t("entry")}</TableHead>
                <TableHead>{t("mark")}</TableHead>
                <TableHead>{t("leverage")}</TableHead>
                <TableHead>{t("pnl")}</TableHead>
                <TableHead>{t("liqPrice")}</TableHead>
                <TableHead>{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {positions.map((p) => (
                <TableRow key={p.id}>
                  <PositionRow position={p} />
                  <td className="px-3 py-2">
                    <div className="flex items-center gap-1">
                      <UpdateTPSLDialog
                        positionId={p.id}
                        currentSL={p.stop_loss}
                        currentTP={p.take_profit}
                      />
                      <ClosePositionButton
                        positionId={p.id}
                        symbol={p.symbol}
                      />
                    </div>
                  </td>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
