"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useClosePosition } from "@/entities/position/model/use-positions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { formatPnL } from "@/shared/lib/format";

const PERCENT_SHORTCUTS = [25, 50, 75, 100] as const;

export function ClosePositionButton({
  positionId,
  symbol,
  quantity,
  initialMargin,
}: {
  positionId: number;
  symbol: string;
  quantity: string;
  initialMargin?: string;
}) {
  const t = useTranslations("position");
  const [open, setOpen] = useState(false);
  const [closeQty, setCloseQty] = useState(quantity);
  const close = useClosePosition();

  const posQty = parseFloat(quantity);

  const handlePercentClick = (pct: number) => {
    if (pct === 100) {
      setCloseQty(quantity);
    } else {
      const val = (posQty * pct) / 100;
      setCloseQty(String(val));
    }
  };

  const handleConfirm = () => {
    const closeQuantity = parseFloat(closeQty);
    if (isNaN(closeQuantity) || closeQuantity <= 0) return;

    const isFullClose = closeQuantity >= posQty;
    const qty = isFullClose ? undefined : closeQty;

    close.mutate(
      { id: positionId, quantity: qty },
      {
        onSuccess: (data) => {
          const pnl = parseFloat(data.realized_pnl);
          const margin = initialMargin ? parseFloat(initialMargin) : 0;
          const pct = margin > 0 ? ((pnl / margin) * 100).toFixed(2) : null;
          const pctStr = pct !== null ? ` (${pnl >= 0 ? "+" : ""}${pct}%)` : "";
          toast.success(
            `${symbol}: ${t("closed")} PnL ${formatPnL(data.realized_pnl)}${pctStr} USDT`,
          );
          setOpen(false);
        },
        onError: () => toast.error(t("closeFailed")),
      },
    );
  };

  const handleOpenChange = (v: boolean) => {
    setOpen(v);
    if (v) setCloseQty(quantity);
  };

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => handleOpenChange(true)}>
        {t("close")}
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t("closePosition")} {symbol}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <p className="text-muted-foreground text-sm">
              {t("currentQty")}: {quantity}
            </p>
            <div className="space-y-2">
              <Label>{t("closeQuantity")}</Label>
              <Input
                type="number"
                step="any"
                min="0"
                max={quantity}
                value={closeQty}
                onChange={(e) => setCloseQty(e.target.value)}
              />
              <div className="flex gap-2">
                {PERCENT_SHORTCUTS.map((pct) => (
                  <Button
                    key={pct}
                    variant="outline"
                    size="sm"
                    type="button"
                    className="flex-1"
                    onClick={() => handlePercentClick(pct)}
                  >
                    {pct}%
                  </Button>
                ))}
              </div>
            </div>
          </div>
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
