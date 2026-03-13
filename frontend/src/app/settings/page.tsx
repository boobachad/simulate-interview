"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Loader2Icon, RefreshCwIcon } from "lucide-react";
import { api } from "@/lib/api";
import { Navbar } from "@/components/Navbar";
import { StatsDisplay } from "@/components/StatsDisplay";

const settingsSchema = z.object({
  leetcode_username: z.string().min(1, "LeetCode username is required"),
  codeforces_username: z.string().min(1, "Codeforces username is required"),
  default_problem_count: z.number().min(1).max(10),
});

export default function SettingsPage() {
  const router = useRouter();
  const [isSyncing, setIsSyncing] = useState(false);
  const [statsKey, setStatsKey] = useState(0);

  const form = useForm<z.infer<typeof settingsSchema>>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      leetcode_username: "",
      codeforces_username: "",
      default_problem_count: 5,
    },
  });

  useEffect(() => {
    const loadProfile = async () => {
    try {
      const profile = await api.profile.get();
      form.reset({
        leetcode_username: profile.leetcode_username,
        codeforces_username: profile.codeforces_username,
        default_problem_count: profile.default_problem_count || 5,
      });
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to load profile";
      toast.error(errorMessage);
      router.push("/login");
    }
    };
    loadProfile();
  }, [form, router]);

  const onSubmit = async (data: z.infer<typeof settingsSchema>) => {
    try {
      await api.profile.update({
        leetcode_username: data.leetcode_username.trim(),
        codeforces_username: data.codeforces_username.trim(),
        default_problem_count: data.default_problem_count,
      });

      toast.success("Settings saved");
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to update profile";
      toast.error(errorMessage);
    }
  };

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      await api.profile.sync();
      toast.success("Stats synced successfully");
      const profile = await api.profile.get();
      form.reset({
        leetcode_username: profile.leetcode_username,
        codeforces_username: profile.codeforces_username,
        default_problem_count: profile.default_problem_count || 5,
      });
      setStatsKey(prev => prev + 1);
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to sync stats";
      toast.error(errorMessage);
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navbar
        breadcrumbItems={[
          { label: "home", href: "/" },
          { label: "settings" },
        ]}
      />
      <div className="container mx-auto px-12 py-12 max-w-[1800px]">
        <div className="flex items-start justify-between mb-12">
          <div>
            <h1 className="text-4xl font-bold mb-3">Settings</h1>
            <p className="text-lg text-muted-foreground">
              Manage your profile and view statistics
            </p>
          </div>
          <Button
            onClick={handleSync}
            disabled={isSyncing}
            variant="outline"
            size="lg"
          >
            {isSyncing ? (
              <>
                <Loader2Icon className="h-5 w-5 animate-spin" />
                Syncing...
              </>
            ) : (
              <>
                <RefreshCwIcon className="h-5 w-5" />
                Sync Stats
              </>
            )}
          </Button>
        </div>

        <div className="grid gap-12 xl:grid-cols-[400px_1fr]">
          <div>
            <div className="rounded-xl border bg-card p-10">
              <h2 className="text-2xl font-semibold mb-8">Profile</h2>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                <Controller
                  name="leetcode_username"
                  control={form.control}
                  render={({ field, fieldState }) => (
                    <Field data-invalid={fieldState.invalid}>
                      <FieldLabel htmlFor="leetcode" className="text-base">
                        LeetCode
                      </FieldLabel>
                      <Input
                        {...field}
                        id="leetcode"
                        placeholder="username"
                        aria-invalid={fieldState.invalid}
                        className="h-12 text-base"
                      />
                      {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                    </Field>
                  )}
                />

                <Controller
                  name="codeforces_username"
                  control={form.control}
                  render={({ field, fieldState }) => (
                    <Field data-invalid={fieldState.invalid}>
                      <FieldLabel htmlFor="codeforces" className="text-base">
                        Codeforces
                      </FieldLabel>
                      <Input
                        {...field}
                        id="codeforces"
                        placeholder="username"
                        aria-invalid={fieldState.invalid}
                        className="h-12 text-base"
                      />
                      {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                    </Field>
                  )}
                />

                <Controller
                  name="default_problem_count"
                  control={form.control}
                  render={({ field, fieldState }) => (
                    <Field data-invalid={fieldState.invalid}>
                      <FieldLabel htmlFor="problem-count" className="text-base">
                        Problems per session
                      </FieldLabel>
                      <select
                        {...field}
                        id="problem-count"
                        className="w-full h-12 px-4 rounded-md border bg-background text-base"
                        onChange={(e) => field.onChange(Number(e.target.value))}
                      >
                        {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((n) => (
                          <option key={n} value={n}>
                            {n}
                          </option>
                        ))}
                      </select>
                      {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                    </Field>
                  )}
                />

                <Button
                  type="submit"
                  disabled={form.formState.isSubmitting}
                  className="w-full h-12 text-base"
                >
                  {form.formState.isSubmitting ? (
                    <>
                      <Loader2Icon className="h-5 w-5 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    "Save Changes"
                  )}
                </Button>
              </form>
            </div>
          </div>

          <div>
            <StatsDisplay key={statsKey} />
          </div>
        </div>
      </div>
    </div>
  );
}
