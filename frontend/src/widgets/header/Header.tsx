"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { useAccount } from "@/entities/account/model/use-account";
import { LogoutButton } from "@/features/auth/logout/LogoutButton";
import { LocaleSwitcher } from "@/features/locale-switcher/LocaleSwitcher";
import { formatUSD } from "@/shared/lib/format";
import { TrendingUp } from "lucide-react";

export function Header() {
  const t = useTranslations("nav");
  const { data: account } = useAccount();

  return (
    <header className="border-b">
      <div className="flex h-14 items-center justify-between px-4">
        <div className="flex items-center gap-6">
          <Link href="/dashboard" className="flex items-center gap-2 font-semibold">
            <TrendingUp className="h-5 w-5" />
            Trading Sim
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link href="/dashboard" className="text-muted-foreground hover:text-foreground transition-colors">
              {t("dashboard")}
            </Link>
            <Link href="/orders" className="text-muted-foreground hover:text-foreground transition-colors">
              {t("orders")}
            </Link>
            <Link href="/trades" className="text-muted-foreground hover:text-foreground transition-colors">
              {t("trades")}
            </Link>
            <Link href="/account" className="text-muted-foreground hover:text-foreground transition-colors">
              {t("account")}
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          {account && (
            <span className="text-sm font-mono">
              {formatUSD(account.equity)} USDT
            </span>
          )}
          <LocaleSwitcher />
          <LogoutButton />
        </div>
      </div>
    </header>
  );
}
