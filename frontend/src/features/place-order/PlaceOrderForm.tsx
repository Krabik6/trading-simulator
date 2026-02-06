"use client";

import { useForm, useWatch } from "react-hook-form";
import { z } from "zod/v4";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useCreateOrder } from "@/entities/order/model/use-orders";
import { useActiveSymbol } from "@/features/symbol-selector/SymbolSelector";
import { usePriceStore } from "@/entities/price/model/price-store";
import { useAccount } from "@/entities/account/model/use-account";
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
import { useState, useMemo } from "react";
import type { OrderSide, OrderType } from "@/entities/order/model/types";
import { formatCrypto, formatUSD } from "@/shared/lib/format";

const schema = z.object({
  quantity: z.string().min(1),
  price: z.string().optional(),
  leverage: z.number().min(1).max(100),
  sizePercent: z.number().min(0).max(100),
  stop_loss: z.string().optional(),
  take_profit: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

const PERCENT_PRESETS = [10, 25, 50, 75, 100];

export function PlaceOrderForm() {
  const t = useTranslations("order");
  const symbol = useActiveSymbol((s) => s.symbol);
  const price = usePriceStore((s) => s.prices[symbol]);
  const { data: account } = useAccount();
  const createOrder = useCreateOrder();
  const [side, setSide] = useState<OrderSide>("BUY");
  const [orderType, setOrderType] = useState<OrderType>("MARKET");

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      quantity: "",
      price: "",
      leverage: 10,
      sizePercent: 0,
      stop_loss: "",
      take_profit: "",
    },
  });

  const leverage = useWatch({ control: form.control, name: "leverage" });
  const sizePercent = useWatch({ control: form.control, name: "sizePercent" });

  const availableMargin = account ? parseFloat(account.available_margin) : 0;
  const currentPrice = price?.mid ?? 0;

  const calculatedQty = useMemo(() => {
    if (!currentPrice || !availableMargin || !sizePercent) return 0;
    return (availableMargin * (sizePercent / 100) * leverage) / currentPrice;
  }, [availableMargin, currentPrice, sizePercent, leverage]);

  const marginUsed = useMemo(() => {
    return availableMargin * (sizePercent / 100);
  }, [availableMargin, sizePercent]);

  const updateQtyFromPercent = (pct: number) => {
    form.setValue("sizePercent", pct);
    if (currentPrice && availableMargin) {
      const qty = (availableMargin * (pct / 100) * leverage) / currentPrice;
      form.setValue("quantity", qty > 0 ? qty.toFixed(6) : "");
    }
  };

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

  const baseCurrency = symbol.replace("USDT", "");

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <div className="flex gap-2">
          <Button
            type="button"
            variant={side === "BUY" ? "default" : "outline"}
            className={cn(
              "flex-1",
              side === "BUY" && "bg-profit hover:bg-profit/90",
            )}
            onClick={() => setSide("BUY")}
          >
            {t("buy")}
          </Button>
          <Button
            type="button"
            variant={side === "SELL" ? "default" : "outline"}
            className={cn(
              "flex-1",
              side === "SELL" && "bg-loss hover:bg-loss/90",
            )}
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
                  onValueChange={([v]) => {
                    field.onChange(v);
                    if (sizePercent > 0) updateQtyFromPercent(sizePercent);
                  }}
                />
              </FormControl>
            </FormItem>
          )}
        />

        {/* Size percent slider */}
        <FormField
          control={form.control}
          name="sizePercent"
          render={() => (
            <FormItem>
              <FormLabel>
                {t("size")}: {sizePercent}%
              </FormLabel>
              <FormControl>
                <Slider
                  min={0}
                  max={100}
                  step={1}
                  value={[sizePercent]}
                  onValueChange={([v]) => updateQtyFromPercent(v)}
                />
              </FormControl>
              <div className="flex gap-1">
                {PERCENT_PRESETS.map((pct) => (
                  <Button
                    key={pct}
                    type="button"
                    variant={sizePercent === pct ? "default" : "outline"}
                    size="sm"
                    className="h-6 flex-1 px-0 text-xs"
                    onClick={() => updateQtyFromPercent(pct)}
                  >
                    {pct}%
                  </Button>
                ))}
              </div>
              {sizePercent > 0 && currentPrice > 0 && (
                <div className="text-muted-foreground space-y-0.5 text-xs">
                  <div>
                    ≈ {formatCrypto(calculatedQty, 6)} {baseCurrency}
                  </div>
                  <div>
                    {t("margin")}: {formatUSD(marginUsed)} USDT
                  </div>
                </div>
              )}
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="quantity"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                {t("quantity")} ({baseCurrency})
              </FormLabel>
              <FormControl>
                <Input
                  placeholder="0.001"
                  {...field}
                  onChange={(e) => {
                    field.onChange(e);
                    form.setValue("sizePercent", 0);
                  }}
                />
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
