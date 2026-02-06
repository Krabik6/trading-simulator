"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useClosePosition } from "@/entities/position/model/use-positions";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { formatPnL } from "@/shared/lib/format";

export function ClosePositionButton({
  positionId,
  symbol,
}: {
  positionId: number;
  symbol: string;
}) {
  const t = useTranslations("position");
  const [open, setOpen] = useState(false);
  const close = useClosePosition();

  const handleConfirm = () => {
    close.mutate(positionId, {
      onSuccess: (data) => {
        toast.success(
          `${symbol}: ${t("closed")} PnL ${formatPnL(data.realized_pnl)} USDT`,
        );
        setOpen(false);
      },
      onError: () => toast.error(t("closeFailed")),
    });
  };

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        {t("close")}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t("closePosition")} {symbol}
            </DialogTitle>
          </DialogHeader>
          <p className="text-muted-foreground text-sm">{t("closeConfirm")}</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              {t("no")}
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirm}
              disabled={close.isPending}
            >
              {t("yes")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
