"use client";

import { useForm } from "react-hook-form";
import { z } from "zod/v4";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { login } from "@/entities/user/api/user-api";
import { useAuthStore } from "@/shared/lib/auth-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from "@/components/ui/form";
import { toast } from "sonner";
import { AxiosError } from "axios";
import { EmailInput } from "@/shared/ui/email-input";

const schema = z.object({
  email: z.email(),
  password: z.string().min(6),
});

type FormValues = z.infer<typeof schema>;

export function LoginForm() {
  const t = useTranslations("auth");
  const router = useRouter();
  const authLogin = useAuthStore((s) => s.login);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      authLogin(data.token, data.user_id);
      router.push("/dashboard");
    },
    onError: (err: AxiosError<{ error: string }>) => {
      toast.error(err.response?.data?.error ?? t("loginFailed"));
    },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit((v) => mutation.mutate(v))}
        className="space-y-4"
      >
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("email")}</FormLabel>
              <FormControl>
                <EmailInput
                  placeholder="user@example.com"
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("password")}</FormLabel>
              <FormControl>
                <Input type="password" placeholder="******" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button
          type="submit"
          className="w-full"
          disabled={mutation.isPending}
        >
          {mutation.isPending ? t("loading") : t("login")}
        </Button>
      </form>
    </Form>
  );
}
