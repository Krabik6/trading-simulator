"use client";

import { useTranslations } from "next-intl";
import { AccountSummary } from "@/widgets/account-summary/AccountSummary";
import { useUser } from "@/entities/user/model/use-user";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDateTime } from "@/shared/lib/format";

export default function AccountPage() {
  const t = useTranslations("account");
  const { data: user } = useUser();

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <AccountSummary />
      {user && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{t("profile")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("email")}</span>
              <span>{user.email}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">{t("createdAt")}</span>
              <span>{formatDateTime(user.created_at)}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
