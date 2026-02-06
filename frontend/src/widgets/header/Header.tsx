"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useLiveEquity } from "@/entities/account/model/use-live-equity";
import { LogoutButton } from "@/features/auth/logout/LogoutButton";
import { LocaleSwitcher } from "@/features/locale-switcher/LocaleSwitcher";
import { formatUSD } from "@/shared/lib/format";
import { cn } from "@/lib/utils";
import { TrendingUp } from "lucide-react";

const navItems = [
  { href: "/dashboard", key: "dashboard" },
  { href: "/orders", key: "orders" },
  { href: "/trades", key: "trades" },
  { href: "/account", key: "account" },
] as const;

export function Header() {
  const t = useTranslations("nav");
  const equity = useLiveEquity();
  const pathname = usePathname();

  return (
    <header className="border-b">
      <div className="flex h-14 items-center justify-between px-4">
        <div className="flex items-center gap-6">
          <Link href="/dashboard" className="flex items-center gap-2 font-semibold">
            <TrendingUp className="h-5 w-5" />
            Trading Sim
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            {navItems.map(({ href, key }) => {
              const active = pathname.endsWith(href);
              return (
                <Link
                  key={key}
                  href={href}
                  className={cn(
                    "transition-colors",
                    active
                      ? "text-foreground font-medium"
                      : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {t(key)}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          {equity !== null && (
            <span className="text-sm font-mono">
              {formatUSD(equity)} USDT
            </span>
          )}
          <LocaleSwitcher />
          <LogoutButton />
        </div>
      </div>
    </header>
  );
}
