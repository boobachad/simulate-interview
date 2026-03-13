"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldError, FieldLabel, FieldDescription } from "@/components/ui/field";
import { Loader2Icon, UserIcon } from "lucide-react";
import { api } from "@/lib/api";

const profileSchema = z.object({
  leetcode_username: z.string().min(1, "LeetCode username is required"),
  codeforces_username: z.string().min(1, "Codeforces username is required"),
});

export default function ProfileSetupPage() {
  const router = useRouter();

  const form = useForm<z.infer<typeof profileSchema>>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      leetcode_username: "",
      codeforces_username: "",
    },
  });

  const onSubmit = async (data: z.infer<typeof profileSchema>) => {
    try {
      await api.profile.setup({
        leetcode_username: data.leetcode_username.trim(),
        codeforces_username: data.codeforces_username.trim(),
      });

      toast.success("Profile created successfully");
      router.push("/start");
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to create profile";
      toast.error(errorMessage);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">Setup Your Profile</h1>
          <p className="text-muted-foreground">
            Connect your coding profiles
          </p>
        </div>

        <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-6">
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <Controller
              name="leetcode_username"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="profile-leetcode">LeetCode Username</FieldLabel>
                  <Input
                    {...field}
                    id="profile-leetcode"
                    type="text"
                    placeholder="your_leetcode_username"
                    aria-invalid={fieldState.invalid}
                    autoComplete="off"
                    autoFocus
                  />
                  <FieldDescription>
                    LeetCode handle
                  </FieldDescription>
                  {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                </Field>
              )}
            />

            <Controller
              name="codeforces_username"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="profile-codeforces">Codeforces Username</FieldLabel>
                  <Input
                    {...field}
                    id="profile-codeforces"
                    type="text"
                    placeholder="your_codeforces_username"
                    aria-invalid={fieldState.invalid}
                    autoComplete="off"
                  />
                  <FieldDescription>
                    Codeforces handle
                  </FieldDescription>
                  {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                </Field>
              )}
            />

            <Button
              type="submit"
              className="w-full"
              disabled={form.formState.isSubmitting}
            >
              {form.formState.isSubmitting ? (
                <>
                  <Loader2Icon className="animate-spin" />
                  Setting up profile...
                </>
              ) : (
                <>
                  <UserIcon />
                  Complete Setup
                </>
              )}
            </Button>
          </form>
        </div>

        <p className="text-center text-sm text-muted-foreground">
          We'll sync your stats
        </p>
      </div>
    </div>
  );
}
