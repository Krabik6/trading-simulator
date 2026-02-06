"use client";

import { PriceTicker } from "@/widgets/price-ticker/PriceTicker";
import { TradingChart } from "@/widgets/trading-chart/TradingChart";
import { OrderPanel } from "@/widgets/order-panel/OrderPanel";
import { PositionsTable } from "@/widgets/positions-table/PositionsTable";
import { OpenOrdersTable } from "@/widgets/open-orders-table/OpenOrdersTable";
import { AccountSummary } from "@/widgets/account-summary/AccountSummary";

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-4 p-4">
      <PriceTicker />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-12">
        <div className="flex flex-col gap-4 lg:col-span-8">
          <div className="bg-card rounded-lg border p-4" style={{ height: 400 }}>
            <TradingChart />
          </div>
          <PositionsTable />
          <OpenOrdersTable />
        </div>
        <div className="flex flex-col gap-4 lg:col-span-4">
          <OrderPanel />
          <AccountSummary />
        </div>
      </div>
    </div>
  );
}
