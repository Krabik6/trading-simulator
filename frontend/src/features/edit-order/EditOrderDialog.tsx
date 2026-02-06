"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod/v4";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useUpdateOrder } from "@/entities/order/model/use-orders";
import type { Order } from "@/entities/order/model/types";
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
import { Pencil } from "lucide-react";

const schema = z.object({
  price: z.string().min(1),
  quantity: z.string().min(1),
  stop_loss: z.string().optional(),
  take_profit: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export function EditOrderDialog({ order }: { order: Order }) {
  const t = useTranslations("order");
  const [open, setOpen] = useState(false);
  const update = useUpdateOrder();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      price: order.price,
      quantity: order.quantity,
      stop_loss: order.stop_loss ?? "",
      take_profit: order.take_profit ?? "",
    },
  });

  const onSubmit = (v: FormValues) => {
    update.mutate(
      {
        id: order.id,
        price: v.price !== order.price ? v.price : undefined,
        quantity: v.quantity !== order.quantity ? v.quantity : undefined,
        stop_loss: v.stop_loss || undefined,
        take_profit: v.take_profit || undefined,
      },
      {
        onSuccess: () => {
          toast.success(t("orderUpdated"));
          setOpen(false);
        },
        onError: () => toast.error(t("updateFailed")),
      },
    );
  };

  return (
    <>
      <Button variant="ghost" size="icon" onClick={() => setOpen(true)}>
        <Pencil className="h-4 w-4" />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("editOrder")}</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="price"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("price")}</FormLabel>
                    <FormControl>
                      <Input type="number" step="any" {...field} />
                    </FormControl>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="quantity"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("quantity")}</FormLabel>
                    <FormControl>
                      <Input type="number" step="any" {...field} />
                    </FormControl>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="stop_loss"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("stopLoss")}</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        step="any"
                        placeholder={t("optional")}
                        {...field}
                      />
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
                      <Input
                        type="number"
                        step="any"
                        placeholder={t("optional")}
                        {...field}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setOpen(false)}
                  type="button"
                >
                  {t("no")}
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
