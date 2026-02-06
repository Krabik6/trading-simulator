"use client";

import { useForm, useWatch } from "react-hook-form";
import { z } from "zod/v4";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useCreateOrder } from "@/entities/order/model/use-orders";
import { useActiveSymbol } from "@/features/symbol-selector/SymbolSelector";
import { usePriceStore } from "@/entities/price/model/price-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Slider } from "@/components/ui/slider";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form";
import { toast } from "sonner";
import axios from "axios";
import { cn } from "@/lib/utils";
import { useState } from "react";
import type { OrderSide, OrderType } from "@/entities/order/model/types";

const schema = z.object({
  quantity: z.string().min(1),
  price: z.string().optional(),
  leverage: z.number().min(1).max(100),
  stop_loss: z.string().optional(),
  take_profit: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export function PlaceOrderForm() {
  const t = useTranslations("order");
  const symbol = useActiveSymbol((s) => s.symbol);
  const price = usePriceStore((s) => s.prices[symbol]);
  const createOrder = useCreateOrder();
  const [side, setSide] = useState<OrderSide>("BUY");
  const [orderType, setOrderType] = useState<OrderType>("MARKET");

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      quantity: "",
      price: "",
      leverage: 10,
      stop_loss: "",
      take_profit: "",
    },
  });

  const onSubmit = (v: FormValues) => {
    createOrder.mutate(
      {
        symbol,
        side,
        type: orderType,
        quantity: v.quantity,
        price: orderType === "LIMIT" ? v.price : undefined,
        leverage: v.leverage,
        stop_loss: v.stop_loss || undefined,
        take_profit: v.take_profit || undefined,
      },
      {
        onSuccess: () => {
          toast.success(t("orderPlaced"));
          form.reset();
        },
        onError: (err) => {
          const msg =
            axios.isAxiosError(err) && err.response?.data?.error
              ? err.response.data.error
              : t("orderFailed");
          toast.error(msg);
        },
      },
    );
  };

  const leverage = useWatch({ control: form.control, name: "leverage" });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <div className="flex gap-2">
          <Button
            type="button"
            variant={side === "BUY" ? "default" : "outline"}
            className={cn("flex-1", side === "BUY" && "bg-profit hover:bg-profit/90")}
            onClick={() => setSide("BUY")}
          >
            {t("buy")}
          </Button>
          <Button
            type="button"
            variant={side === "SELL" ? "default" : "outline"}
            className={cn("flex-1", side === "SELL" && "bg-loss hover:bg-loss/90")}
            onClick={() => setSide("SELL")}
          >
            {t("sell")}
          </Button>
        </div>

        <Tabs
          value={orderType}
          onValueChange={(v) => setOrderType(v as OrderType)}
        >
          <TabsList className="w-full">
            <TabsTrigger value="MARKET" className="flex-1">
              {t("market")}
            </TabsTrigger>
            <TabsTrigger value="LIMIT" className="flex-1">
              {t("limit")}
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {price && (
          <div className="text-muted-foreground text-center text-xs">
            {t("currentPrice")}: {price.mid.toFixed(2)}
          </div>
        )}

        <FormField
          control={form.control}
          name="quantity"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("quantity")}</FormLabel>
              <FormControl>
                <Input placeholder="0.01" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {orderType === "LIMIT" && (
          <FormField
            control={form.control}
            name="price"
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t("price")}</FormLabel>
                <FormControl>
                  <Input placeholder="50000" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )}

        <FormField
          control={form.control}
          name="leverage"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                {t("leverage")}: {leverage}x
              </FormLabel>
              <FormControl>
                <Slider
                  min={1}
                  max={100}
                  step={1}
                  value={[field.value]}
                  onValueChange={([v]) => field.onChange(v)}
                />
              </FormControl>
            </FormItem>
          )}
        />

        <div className="grid grid-cols-2 gap-2">
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
        </div>

        <Button
          type="submit"
          className={cn(
            "w-full",
            side === "BUY"
              ? "bg-profit hover:bg-profit/90"
              : "bg-loss hover:bg-loss/90",
          )}
          disabled={createOrder.isPending}
        >
          {createOrder.isPending
            ? t("placing")
            : `${side === "BUY" ? t("buy") : t("sell")} ${symbol}`}
        </Button>
      </form>
    </Form>
  );
}
