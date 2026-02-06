"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { LoginForm } from "@/features/auth/login/LoginForm";
import { useAuthStore } from "@/shared/lib/auth-store";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TrendingUp } from "lucide-react";

export default function LoginPage() {
  const t = useTranslations("auth");
  const router = useRouter();
  const { isAuthenticated, hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  useEffect(() => {
    if (isAuthenticated) router.push("/dashboard");
  }, [isAuthenticated, router]);

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <TrendingUp className="h-5 w-5" />
          </div>
          <CardTitle>{t("login")}</CardTitle>
          <CardDescription>Trading Simulator</CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm />
          <p className="text-muted-foreground mt-4 text-center text-sm">
            {t("noAccount")}{" "}
            <Link href="/register" className="text-primary underline">
              {t("signUpLink")}
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
