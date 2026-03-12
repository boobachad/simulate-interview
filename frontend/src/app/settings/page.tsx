"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldError, FieldLabel, FieldDescription } from "@/components/ui/field";
import { Loader2Icon, RefreshCwIcon, SaveIcon } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { Navbar } from "@/components/Navbar";

const settingsSchema = z.object({
  leetcode_username: z.string().min(1, "LeetCode username is required"),
  codeforces_username: z.string().min(1, "Codeforces username is required"),
});

export default function SettingsPage() {
  const router = useRouter();

  const form = useForm<z.infer<typeof settingsSchema>>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      leetcode_username: "",
      codeforces_username: "",
    },
  });

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const profile = await apiClient.profile.get();
      form.reset({
        leetcode_username: profile.leetcode_username,
        codeforces_username: profile.codeforces_username,
      });
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to load profile";
      toast.error(errorMessage);
      router.push("/login");
    }
  };

  const onSubmit = async (data: z.infer<typeof settingsSchema>) => {
    try {
      await apiClient.profile.update({
        leetcode_username: data.leetcode_username.trim(),
        codeforces_username: data.codeforces_username.trim(),
      });

      toast.success("Profile updated successfully");
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to update profile";
      toast.error(errorMessage);
    }
  };

  const handleSync = async () => {
    try {
      const response = await apiClient.profile.sync();
      toast.success(`Stats synced at ${response.synced_at}`);
      await loadProfile();
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to sync stats";
      toast.error(errorMessage);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar
        breadcrumbItems={[
          { label: "home", href: "/" },
          { label: "settings" },
        ]}
      />
      <div className="flex-1 w-full p-6 md:p-12">
        <div className="max-w-2xl mx-auto space-y-8">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
            <p className="text-muted-foreground mt-1">
              Manage your profile
            </p>
          </div>

          <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-6 space-y-6">
            <div>
              <h2 className="text-lg font-semibold">Coding Profiles</h2>
              <p className="text-sm text-muted-foreground">
                Update your LeetCode and Codeforces handles
              </p>
            </div>

            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <Controller
                name="leetcode_username"
                control={form.control}
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="settings-leetcode">
                      LeetCode handle
                    </FieldLabel>
                    <Input
                      {...field}
                      id="settings-leetcode"
                      type="text"
                      placeholder="your_leetcode_username"
                      aria-invalid={fieldState.invalid}
                      autoComplete="off"
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
                    <FieldLabel htmlFor="settings-codeforces">
                      Codeforces handle
                    </FieldLabel>
                    <Input
                      {...field}
                      id="settings-codeforces"
                      type="text"
                      placeholder="your_codeforces_username"
                      aria-invalid={fieldState.invalid}
                      autoComplete="off"
                    />
                    {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                  </Field>
                )}
              />

              <div className="flex gap-2">
                <Button
                  type="submit"
                  disabled={form.formState.isSubmitting}
                  className="flex-1"
                >
                  {form.formState.isSubmitting ? (
                    <>
                      <Loader2Icon className="animate-spin" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <SaveIcon />
                      Save Changes
                    </>
                  )}
                </Button>

                <Button
                  type="button"
                  variant="outline"
                  onClick={handleSync}
                  disabled={form.formState.isSubmitting}
                >
                  <RefreshCwIcon />
                  Sync Stats
                </Button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
