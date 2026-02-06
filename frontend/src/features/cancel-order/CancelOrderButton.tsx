"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useCancelOrder } from "@/entities/order/model/use-orders";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { X } from "lucide-react";

export function CancelOrderButton({ orderId }: { orderId: number }) {
  const t = useTranslations("order");
  const [open, setOpen] = useState(false);
  const cancel = useCancelOrder();

  const handleConfirm = () => {
    cancel.mutate(orderId, {
      onSuccess: () => {
        toast.success(t("orderCancelled"));
        setOpen(false);
      },
      onError: () => toast.error(t("cancelFailed")),
    });
  };

  return (
    <>
      <Button variant="ghost" size="icon" onClick={() => setOpen(true)}>
        <X className="h-4 w-4" />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("cancelOrder")}</DialogTitle>
          </DialogHeader>
          <p className="text-muted-foreground text-sm">{t("cancelConfirm")}</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              {t("no")}
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirm}
              disabled={cancel.isPending}
            >
              {t("yes")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
