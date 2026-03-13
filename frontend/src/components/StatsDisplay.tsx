"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { statsCache } from "@/lib/dexie-db";
import { toUserID } from "@/lib/api";
import type { CombinedStats } from "@/lib/api";
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

        if (userId) {
          const cached = await statsCache.get(toUserID(userId));
          if (cached && !controller.signal.aborted) {
            setStats(cached);
            setCachedAt(cached.cached_at);
            setLoading(false);
            return;
          }
        }

        const data = await api.stats.get(controller.signal);
        
        if (controller.signal.aborted) return;

        setStats(data);
        setCachedAt(data.cached_at);

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

  const leetcodeSkills = stats.leetcode?.skills 
    ? Object.entries(stats.leetcode.skills).sort((a, b) => b[1].problem_count - a[1].problem_count)
    : [];

  const codeforcesTagsArray = stats.codeforces?.tags
    ? Object.entries(stats.codeforces.tags).sort((a, b) => b[1] - a[1])
    : [];

  // Combine all topics and find weakest 10
  const allTopics: Array<{ name: string; count: number; platform: string }> = [];
  
  if (stats.leetcode?.skills) {
    Object.entries(stats.leetcode.skills).forEach(([tag, skill]) => {
      allTopics.push({ name: tag, count: skill.problem_count, platform: 'LeetCode' });
    });
  }
  
  if (stats.codeforces?.tags) {
    Object.entries(stats.codeforces.tags).forEach(([tag, count]) => {
      allTopics.push({ name: tag, count, platform: 'Codeforces' });
    });
  }
  
  const weakestTopics = allTopics
    .sort((a, b) => a.count - b.count)
    .slice(0, 10)
    .map(t => ({ name: t.name.toLowerCase(), platform: t.platform }));

  const isWeakTopic = (topicName: string, platform: string) => 
    weakestTopics.some(wt => wt.name === topicName.toLowerCase() && wt.platform === platform);

  return (
    <div className={className}>
      <div className="grid gap-8 lg:grid-cols-2">
        {hasLeetCode && stats.leetcode ? (
          <Card className="rounded-xl">
            <CardHeader className="pb-6">
              <div className="flex items-center gap-3 mb-2">
                <Code2 className="h-6 w-6 text-primary" />
                <CardTitle className="text-2xl">LeetCode</CardTitle>
              </div>
              <CardDescription className="text-base">@{stats.leetcode.username}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-8">
              <div>
                <div className="flex items-baseline gap-3 mb-4">
                  <span className="text-5xl font-bold">
                    {stats.leetcode.total_solved}
                  </span>
                  <span className="text-lg text-muted-foreground">
                    problems solved
                  </span>
                </div>
                <div className="flex flex-wrap gap-3">
                  <Badge variant="outline" className="text-base px-4 py-2 text-green-600 border-green-600/30">
                    Easy: {stats.leetcode.easy_solved}
                  </Badge>
                  <Badge variant="outline" className="text-base px-4 py-2 text-yellow-600 border-yellow-600/30">
                    Medium: {stats.leetcode.medium_solved}
                  </Badge>
                  <Badge variant="outline" className="text-base px-4 py-2 text-red-600 border-red-600/30">
                    Hard: {stats.leetcode.hard_solved}
                  </Badge>
                </div>
              </div>

              {leetcodeSkills.length > 0 && (
                <div className="space-y-4">
                  <div className="text-base font-semibold">
                    Skills by Topic
                  </div>
                  <div className="max-h-[600px] overflow-y-auto pr-2">
                    <div className="flex flex-wrap gap-2">
                      {leetcodeSkills.map(([tag, skill]) => (
                        <Badge
                          key={tag}
                          variant="outline"
                          className={`px-4 py-2 text-sm font-medium ${
                            isWeakTopic(tag, 'LeetCode')
                              ? 'border-orange-500/50 bg-orange-50/10 text-orange-600'
                              : ''
                          }`}
                        >
                          {tag} ×{skill.problem_count}
                          <span className="ml-2 text-xs opacity-70">
                            {skill.level}
                          </span>
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        ) : (
          <Card className="border-dashed rounded-xl">
            <CardHeader className="pb-6">
              <div className="flex items-center gap-3 mb-2">
                <Code2 className="h-6 w-6 text-muted-foreground" />
                <CardTitle className="text-2xl text-muted-foreground">
                  LeetCode
                </CardTitle>
              </div>
              <CardDescription className="text-base">No data available</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-base text-muted-foreground">
                LeetCode statistics could not be fetched. Check your username
                in settings.
              </p>
            </CardContent>
          </Card>
        )}

        {hasCodeforces && stats.codeforces ? (
          <Card className="rounded-xl">
            <CardHeader className="pb-6">
              <div className="flex items-center gap-3 mb-2">
                <Award className="h-6 w-6 text-primary" />
                <CardTitle className="text-2xl">Codeforces</CardTitle>
              </div>
              <CardDescription className="text-base">@{stats.codeforces.username}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-8">
              <div>
                <div className="flex items-center gap-6 mb-4">
                  <div>
                    <div className="flex items-baseline gap-3">
                      <span className="text-5xl font-bold">
                        {stats.codeforces.rating}
                      </span>
                      <span className="text-lg text-muted-foreground">
                        rating
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 text-base text-muted-foreground">
                    <TrendingUp className="h-5 w-5" />
                    <span>Max: {stats.codeforces.max_rating}</span>
                  </div>
                </div>
                <Badge variant="outline" className="text-base px-4 py-2">
                  {stats.codeforces.problems_solved} problems solved
                </Badge>
              </div>

              {codeforcesTagsArray.length > 0 && (
                <div className="space-y-4">
                  <div className="text-base font-semibold">
                    Problem Tags
                  </div>
                  <div className="max-h-[600px] overflow-y-auto pr-2">
                    <div className="flex flex-wrap gap-2">
                      {codeforcesTagsArray.map(([tag, count]) => (
                        <Badge
                          key={tag}
                          variant="outline"
                          className={`px-4 py-2 text-sm font-medium ${
                            isWeakTopic(tag, 'Codeforces')
                              ? 'border-orange-500/50 bg-orange-50/10 text-orange-600'
                              : ''
                          }`}
                        >
                          {tag} ×{count}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        ) : (
          <Card className="border-dashed rounded-xl">
            <CardHeader className="pb-6">
              <div className="flex items-center gap-3 mb-2">
                <Award className="h-6 w-6 text-muted-foreground" />
                <CardTitle className="text-2xl text-muted-foreground">
                  Codeforces
                </CardTitle>
              </div>
              <CardDescription className="text-base">No data available</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-base text-muted-foreground">
                Codeforces statistics could not be fetched. Check your username
                in settings.
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {cachedAt && (
        <p className="mt-6 text-sm text-muted-foreground text-center">
          Last updated: {new Date(cachedAt).toLocaleString()}
        </p>
      )}
    </div>
  );
}
