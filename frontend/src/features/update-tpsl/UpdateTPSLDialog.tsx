"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod/v4";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useUpdateTPSL } from "@/entities/position/model/use-positions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { Settings2 } from "lucide-react";

const PERCENT_OPTIONS = [25, 50, 75, 100];

const schema = z.object({
  stop_loss: z.string().optional(),
  take_profit: z.string().optional(),
  sl_close_percent: z.string(),
  tp_close_percent: z.string(),
});

type FormValues = z.infer<typeof schema>;

export function UpdateTPSLDialog({
  positionId,
  currentSL,
  currentTP,
  currentSLPercent,
  currentTPPercent,
}: {
  positionId: number;
  currentSL: string | null;
  currentTP: string | null;
  currentSLPercent?: number;
  currentTPPercent?: number;
}) {
  const t = useTranslations("position");
  const [open, setOpen] = useState(false);
  const update = useUpdateTPSL();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      stop_loss: currentSL ?? "",
      take_profit: currentTP ?? "",
      sl_close_percent: String(currentSLPercent ?? 100),
      tp_close_percent: String(currentTPPercent ?? 100),
    },
  });

  const onSubmit = (v: FormValues) => {
    update.mutate(
      {
        id: positionId,
        stop_loss: v.stop_loss || undefined,
        take_profit: v.take_profit || undefined,
        sl_close_percent: parseInt(v.sl_close_percent),
        tp_close_percent: parseInt(v.tp_close_percent),
      },
      {
        onSuccess: () => {
          toast.success(t("tpslUpdated"));
          setOpen(false);
        },
        onError: () => toast.error(t("tpslFailed")),
      },
    );
  };

  return (
    <>
      <Button variant="ghost" size="icon" onClick={() => setOpen(true)}>
        <Settings2 className="h-4 w-4" />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("updateTPSL")}</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="stop_loss"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("stopLoss")}</FormLabel>
                    <FormControl>
                      <Input placeholder={t("optional")} {...field} />
                    </FormControl>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="sl_close_percent"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("slClosePercent")}</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PERCENT_OPTIONS.map((pct) => (
                          <SelectItem key={pct} value={String(pct)}>
                            {pct}%
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="take_profit"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("takeProfit")}</FormLabel>
                    <FormControl>
                      <Input placeholder={t("optional")} {...field} />
                    </FormControl>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="tp_close_percent"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("tpClosePercent")}</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PERCENT_OPTIONS.map((pct) => (
                          <SelectItem key={pct} value={String(pct)}>
                            {pct}%
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" onClick={() => setOpen(false)} type="button">
                  {t("cancel")}
                </Button>
                <Button type="submit" disabled={update.isPending}>
                  {t("save")}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </>
  );
}
