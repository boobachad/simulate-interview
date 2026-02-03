'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Navbar } from '@/components/Navbar';
import { statsApi, UserProfile } from '@/lib/api';

export default function Onboarding() {
  const router = useRouter();

  const [name, setName] = useState('');
  const [leetcodeUsername, setLeetcodeUsername] = useState('');
  const [codeforcesUsername, setCodeforcesUsername] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [error, setError] = useState('');

  const handleGetStats = async () => {
    if (!name.trim()) {
      setError('Please enter your name');
      return;
    }

    if (!leetcodeUsername.trim() && !codeforcesUsername.trim()) {
      setError('Please enter at least one username');
      return;
    }

    setError('');
    setIsLoading(true);

    try {
      const profile = await statsApi.getUserStats(
        name,
        leetcodeUsername || undefined,
        codeforcesUsername || undefined
      );
      setUserProfile(profile);
    } catch (err: any) {
      console.error('Error fetching stats:', err);
      setError(err.response?.data?.error || 'Failed to fetch stats. Please check your usernames and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleContinue = () => {
    // Store user profile in localStorage or state if needed
    if (userProfile && userProfile.suggested_areas && userProfile.suggested_areas.length > 0) {
      // Could pre-select suggested areas in the start page
      router.push('/start');
    } else {
      router.push('/start');
    }
  };

  return (
    <div className="min-h-screen bg-bg-primary">
      <Navbar breadcrumbItems={[{ label: 'home', href: '/' }, { label: 'onboarding' }]} />
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-4xl mx-auto">
          {/* Header */}
          <div className="text-center mb-12">
            <h1 className="text-5xl font-bold text-content-primary mb-4">Welcome</h1>
          </div>

          {!userProfile ? (
            /* Input Form */
            <div className="bg-surface-elevated rounded-lg shadow-2xl p-8 border border-border-primary">
              <h2 className="text-2xl font-bold mb-6 text-content-primary">Tell us about yourself</h2>

              <div className="space-y-6">
                {/* Name Input */}
                <div>
                  <label className="block text-sm font-medium text-content-primary mb-2">
                    Your Name <span className="text-error">*</span>
                  </label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Enter your name"
                    className="w-full px-4 py-3 bg-surface-base border border-border-secondary rounded-lg text-content-primary placeholder-content-tertiary focus:outline-none focus:ring-2 focus:ring-border-focus"
                    onKeyDown={(e) => e.key === 'Enter' && handleGetStats()}
                  />
                </div>

                {/* LeetCode Username */}
                <div>
                  <label className="block text-sm font-medium text-content-primary mb-2">
                    LeetCode Username
                  </label>
                  <input
                    type="text"
                    value={leetcodeUsername}
                    onChange={(e) => setLeetcodeUsername(e.target.value)}
                    placeholder="Enter your LeetCode username"
                    className="w-full px-4 py-3 bg-surface-base border border-border-secondary rounded-lg text-content-primary placeholder-content-tertiary focus:outline-none focus:ring-2 focus:ring-border-focus"
                  />
                </div>

                {/* Codeforces Username */}
                <div>
                  <label className="block text-sm font-medium text-content-primary mb-2">
                    Codeforces Username
                  </label>
                  <input
                    type="text"
                    value={codeforcesUsername}
                    onChange={(e) => setCodeforcesUsername(e.target.value)}
                    placeholder="Enter your Codeforces username"
                    className="w-full px-4 py-3 bg-surface-base border border-border-secondary rounded-lg text-content-primary placeholder-content-tertiary focus:outline-none focus:ring-2 focus:ring-border-focus"
                  />
                </div>

                {/* Error Message */}
                {error && (
                  <div className="p-4 bg-error-bg border border-error-border rounded-lg">
                    <p className="text-sm text-error">{error}</p>
                  </div>
                )}

                {/* Get Stats Button */}
                <button
                  onClick={handleGetStats}
                  disabled={isLoading}
                  className="w-full py-4 px-6 bg-primary text-primary-foreground text-lg font-semibold rounded-lg hover:bg-primary/90 disabled:bg-muted disabled:text-muted-foreground disabled:cursor-not-allowed transition-all shadow-lg hover:shadow-xl"
                >
                  {isLoading ? (
                    <span className="flex items-center justify-center">
                      <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-content-inverse mr-3"></div>
                      Fetching Stats...
                    </span>
                  ) : (
                    'Get Stats'
                  )}
                </button>

                {/* Skip Button */}
                <button
                  onClick={() => router.push('/start')}
                  className="w-full py-3 px-6 bg-surface-elevated text-content-primary text-base font-medium rounded-lg hover:bg-surface-hover transition-all border border-border-primary"
                >
                  Skip for now
                </button>
              </div>
            </div>
          ) : (
            /* Stats Display */
            <div className="space-y-6">
              {/* Welcome Message */}
              <div className="bg-surface-elevated rounded-lg p-6 border border-border-primary">
                <h2 className="text-2xl font-bold text-content-primary mb-2">
                  Hello {userProfile.name},
                </h2>
              </div>

              {/* LeetCode Stats */}
              {userProfile.leetcode_stats && (
                <div className="bg-surface-elevated rounded-lg p-6 border border-border-primary">
                  <h3 className="text-xl font-bold text-content-primary mb-4">LeetCode Stats</h3>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                    <div className="text-center">
                      <div className="text-3xl font-bold text-primary">{userProfile.leetcode_stats.total_solved}</div>
                      <div className="text-sm text-content-secondary">Total Solved</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-success">{userProfile.leetcode_stats.easy_solved}</div>
                      <div className="text-sm text-content-secondary">Easy</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-warning">{userProfile.leetcode_stats.medium_solved}</div>
                      <div className="text-sm text-content-secondary">Medium</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-error">{userProfile.leetcode_stats.hard_solved}</div>
                      <div className="text-sm text-content-secondary">Hard</div>
                    </div>
                  </div>

                  {/* Top Skills */}
                  <div className="mt-4">
                    <h4 className="font-semibold text-content-primary mb-2">Top Skills:</h4>
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(userProfile.leetcode_stats.skills)
                        .sort((a, b) => b[1].problem_count - a[1].problem_count)
                        .slice(0, 8)
                        .map(([tag, skill]) => (
                          <span
                            key={tag}
                            className="px-3 py-1 bg-info-bg text-info text-xs font-medium rounded-full"
                          >
                            {tag} ×{skill.problem_count}
                          </span>
                        ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Codeforces Stats */}
              {userProfile.codeforces_stats && (
                <div className="bg-surface-elevated rounded-lg p-6 border border-border-primary">
                  <h3 className="text-xl font-bold text-content-primary mb-4">Codeforces Stats</h3>
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
                    <div className="text-center">
                      <div className="text-3xl font-bold text-primary">{userProfile.codeforces_stats.rating}</div>
                      <div className="text-sm text-content-secondary">Current Rating</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-warning">{userProfile.codeforces_stats.max_rating}</div>
                      <div className="text-sm text-content-secondary">Max Rating</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-success">{userProfile.codeforces_stats.problems_solved}</div>
                      <div className="text-sm text-content-secondary">Problems Solved</div>
                    </div>
                  </div>

                  {/* Rank */}
                  <div className="text-center mb-4">
                    <span className="text-lg font-semibold text-content-primary">
                      Rank: <span className="text-primary">{userProfile.codeforces_stats.rank}</span>
                    </span>
                  </div>

                  {/* Top Tags */}
                  {Object.keys(userProfile.codeforces_stats.tags).length > 0 && (
                    <div className="mt-4">
                      <h4 className="font-semibold text-content-primary mb-2">Top Tags:</h4>
                      <div className="flex flex-wrap gap-2">
                        {Object.entries(userProfile.codeforces_stats.tags)
                          .sort((a, b) => b[1] - a[1])
                          .slice(0, 8)
                          .map(([tag, count]) => (
                            <span
                              key={tag}
                              className="px-3 py-1 bg-info-bg text-info text-xs font-medium rounded-full"
                            >
                              {tag} ×{count}
                            </span>
                          ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Suggested Focus Areas */}
              {userProfile.suggested_areas && userProfile.suggested_areas.length > 0 && (
                <div className="bg-surface-elevated rounded-lg p-6 border border-border-primary">
                  <h3 className="text-xl font-bold text-content-primary mb-4">Suggested Focus Areas</h3>
                  <p className="text-content-secondary mb-4">
                    Based on your stats, we recommend practicing these areas:
                  </p>
                  <div className="flex flex-wrap gap-3">
                    {userProfile.suggested_areas.map((area) => (
                      <span
                        key={area}
                        className="px-4 py-2 bg-primary/10 text-primary font-medium rounded-lg border border-primary"
                      >
                        {area.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Continue Button */}
              <div className="text-center">
                <button
                  onClick={handleContinue}
                  className="py-4 px-12 bg-primary text-primary-foreground text-xl font-semibold rounded-lg hover:bg-primary/90 transition-all shadow-lg hover:shadow-xl"
                >
                  Continue to Practice
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
