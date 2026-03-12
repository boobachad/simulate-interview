"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { statsCache } from "@/lib/dexie-db";
import { toUserID } from "@/types/api";
import type { CombinedStats } from "@/types/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { AlertCircle, TrendingUp, Award, Code2 } from "lucide-react";

interface StatsDisplayProps {
  userId?: string;
  className?: string;
}

export function StatsDisplay({ userId, className }: StatsDisplayProps) {
  const [stats, setStats] = useState<CombinedStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cachedAt, setCachedAt] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    const fetchStats = async () => {
      try {
        setLoading(true);
        setError(null);

        // Try cache first if userId provided
        if (userId) {
          const cached = await statsCache.get(toUserID(userId));
          if (cached && !controller.signal.aborted) {
            setStats(cached);
            setCachedAt(cached.cached_at);
            setLoading(false);
            return;
          }
        }

        // Fetch from API
        const data = await apiClient.stats.get();
        
        if (controller.signal.aborted) return;

        setStats(data);
        setCachedAt(data.cached_at);

        // Cache the result if userId provided
        if (userId) {
          await statsCache.set(toUserID(userId), data);
        }
      } catch (err) {
        if (controller.signal.aborted) return;
        
        setError(
          err instanceof Error ? err.message : "Failed to fetch statistics"
        );
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };

    fetchStats();

    return () => controller.abort();
  }, [userId]);

  if (loading) {
    return (
      <div className={className}>
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-48" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-20 w-full" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-48" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-20 w-full" />
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={className}>
        <Card className="border-destructive">
          <CardContent className="flex items-center gap-2 pt-6">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <p className="text-sm text-destructive">{error}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  const hasLeetCode = stats.leetcode !== null;
  const hasCodeforces = stats.codeforces !== null;

  return (
    <div className={className}>
      <div className="grid gap-4 md:grid-cols-2">
        {/* LeetCode Stats */}
        {hasLeetCode && stats.leetcode ? (
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Code2 className="h-5 w-5 text-primary" />
                <CardTitle>LeetCode</CardTitle>
              </div>
              <CardDescription>@{stats.leetcode.username}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-bold">
                      {stats.leetcode.total_solved}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      problems solved
                    </span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Badge variant="outline" className="text-green-600">
                    Easy: {stats.leetcode.easy_solved}
                  </Badge>
                  <Badge variant="outline" className="text-yellow-600">
                    Medium: {stats.leetcode.medium_solved}
                  </Badge>
                  <Badge variant="outline" className="text-red-600">
                    Hard: {stats.leetcode.hard_solved}
                  </Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        ) : (
          <Card className="border-dashed">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Code2 className="h-5 w-5 text-muted-foreground" />
                <CardTitle className="text-muted-foreground">
                  LeetCode
                </CardTitle>
              </div>
              <CardDescription>No data available</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                LeetCode statistics could not be fetched. Check your username
                in settings.
              </p>
            </CardContent>
          </Card>
        )}

        {/* Codeforces Stats */}
        {hasCodeforces && stats.codeforces ? (
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Award className="h-5 w-5 text-primary" />
                <CardTitle>Codeforces</CardTitle>
              </div>
              <CardDescription>@{stats.codeforces.username}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center gap-4">
                  <div>
                    <div className="flex items-baseline gap-2">
                      <span className="text-3xl font-bold">
                        {stats.codeforces.rating}
                      </span>
                      <span className="text-sm text-muted-foreground">
                        rating
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1 text-sm text-muted-foreground">
                    <TrendingUp className="h-4 w-4" />
                    <span>Max: {stats.codeforces.max_rating}</span>
                  </div>
                </div>
                <div>
                  <Badge variant="outline">
                    {stats.codeforces.problems_solved} problems solved
                  </Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        ) : (
          <Card className="border-dashed">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Award className="h-5 w-5 text-muted-foreground" />
                <CardTitle className="text-muted-foreground">
                  Codeforces
                </CardTitle>
              </div>
              <CardDescription>No data available</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Codeforces statistics could not be fetched. Check your username
                in settings.
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Cache timestamp */}
      {cachedAt && (
        <p className="mt-2 text-xs text-muted-foreground text-center">
          Last updated: {new Date(cachedAt).toLocaleString()}
        </p>
      )}
    </div>
  );
}
