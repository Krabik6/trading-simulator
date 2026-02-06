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
import { toast } from "sonner";
import { Settings2 } from "lucide-react";

const schema = z.object({
  stop_loss: z.string().optional(),
  take_profit: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export function UpdateTPSLDialog({
  positionId,
  currentSL,
  currentTP,
}: {
  positionId: number;
  currentSL: string | null;
  currentTP: string | null;
}) {
  const t = useTranslations("position");
  const [open, setOpen] = useState(false);
  const update = useUpdateTPSL();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      stop_loss: currentSL ?? "",
      take_profit: currentTP ?? "",
    },
  });

  const onSubmit = (v: FormValues) => {
    update.mutate(
      {
        id: positionId,
        stop_loss: v.stop_loss || undefined,
        take_profit: v.take_profit || undefined,
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
