"use client";

import { useEffect, useSyncExternalStore } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/shared/lib/auth-store";
import { WebSocketProvider } from "@/shared/providers/websocket-provider";
import { Header } from "@/widgets/header/Header";

function useHydrated() {
  return useSyncExternalStore(
    (cb) => {
      // On mount, we're hydrated
      cb();
      return () => {};
    },
    () => true,
    () => false,
  );
}

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const hydrated = useHydrated();
  const { isAuthenticated, hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  useEffect(() => {
    if (hydrated && !isAuthenticated) {
      router.push("/login");
    }
  }, [hydrated, isAuthenticated, router]);

  if (!hydrated || !isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-muted-foreground animate-pulse">Loading...</div>
      </div>
    );
  }

  return (
    <WebSocketProvider>
      <div className="flex min-h-screen flex-col">
        <Header />
        <main className="flex-1">{children}</main>
      </div>
    </WebSocketProvider>
  );
}
