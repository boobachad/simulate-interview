// Branded types for domain IDs
type Brand<T, B extends string> = T & { readonly __brand: B };

export type UserID = Brand<string, "UserID">;
export type SessionID = Brand<string, "SessionID">;
export type ProblemID = Brand<string, "ProblemID">;

// constructors
export const toUserID = (id: string): UserID => id as UserID;
export const toSessionID = (id: string): SessionID => id as SessionID;
export const toProblemID = (id: string): ProblemID => id as ProblemID;

// Authentication
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  has_profile: boolean;
}

// Profile
export interface ProfileRequest {
  leetcode_username: string;
  codeforces_username: string;
}

export interface ProfileData {
  leetcode_username: string;
  codeforces_username: string;
  last_synced: string | null;
  problem_count: number;
}

// Stats
export interface LeetCodeStats {
  username: string;
  total_solved: number;
  easy_solved: number;
  medium_solved: number;
  hard_solved: number;
  topics: Record<string, number>;
}

export interface CodeforcesStats {
  username: string;
  rating: number;
  max_rating: number;
  problems_solved: number;
  tags: Record<string, number>;
}

export interface CombinedStats {
  leetcode: LeetCodeStats | null;
  codeforces: CodeforcesStats | null;
  cached_at: string;
}

// Focus Areas
export interface FocusArea {
  platform: "leetcode" | "codeforces";
  topic: string;
  problem_count: number;
  user_solved?: number;
}

// Focus Selection
export type FocusSelection =
  | { mode: "all" }
  | { mode: "single"; topic: string }
  | { mode: "multiple"; topics: string[] };

// Type guard for focus selection
export function isSingleTopicMode(
  selection: FocusSelection
): selection is { mode: "single"; topic: string } {
  return selection.mode === "single";
}

export function isMultipleTopicsMode(
  selection: FocusSelection
): selection is { mode: "multiple"; topics: string[] } {
  return selection.mode === "multiple";
}

// Test Case
export interface TestCase {
  input: string;
  expected_output: string;
  explanation?: string;
}

// Problem
export interface Problem {
  id: ProblemID;
  title: string;
  description: string;
  difficulty?: "easy" | "medium" | "hard";
  focus_area: string;
  sample_cases: TestCase[];
  created_at: string;
}

// Sessions
export type SessionCreateRequest =
  | {
      focus_mode: "all";
      problem_count: number;
    }
  | {
      focus_mode: "single";
      focus_topic: string;
      problem_count: number;
    }
  | {
      focus_mode: "multiple";
      focus_topics: string[];
      problem_count: number;
    };

export interface SessionCreateResponse {
  session_id: SessionID;
  first_problem: Problem;
}

// Session Problem Status
export type SessionProblemStatus = "generating" | "ready" | "failed";

export interface SessionProblem {
  problem_number: number;
  status: SessionProblemStatus;
  problem?: Problem;
  error_message?: string;
}

// Type guard for session problem status
export function isReady(
  problem: SessionProblem
): problem is SessionProblem & { status: "ready"; problem: Problem } {
  return problem.status === "ready" && problem.problem !== undefined;
}

// Exhaustive status handler
export function getStatusMessage(status: SessionProblemStatus): string {
  switch (status) {
    case "generating":
      return "Generating problem...";
    case "ready":
      return "Problem ready";
    case "failed":
      return "Generation failed";
    default: {
      const _exhaustive: never = status;
      throw new Error(`Unhandled status: ${_exhaustive}`);
    }
  }
}

export interface SessionData {
  id: SessionID;
  problem_count: number;
  focus_mode: "all" | "single" | "multiple";
  focus_topic: string | null;
  focus_topics: string[] | null;
  current_problem_number: number;
  status: "active" | "completed";
  problems: SessionProblem[];
}

export interface NextProblemResponse {
  ready: boolean;
  problem?: Problem;
}

// Immutable utility type with proper union handling
export type Immutable<T> = {
  readonly [K in keyof T]: [T[K]] extends [object]
    ? Immutable<Exclude<T[K], null | undefined>> | Extract<T[K], null | undefined>
    : T[K];
};

// API endpoints
export const API_ENDPOINTS = {
  auth: {
    login: "/api/auth/login",
    logout: "/api/auth/logout",
  },
  profile: {
    setup: "/api/profile/setup",
    get: "/api/profile",
    update: "/api/profile",
    sync: "/api/profile/sync",
  },
  stats: {
    get: "/api/stats",
  },
  focusAreas: {
    list: "/api/focus-areas",
  },
  sessions: {
    create: "/api/sessions",
    get: (id: SessionID) => `/api/sessions/${id}`,
    next: (id: SessionID, num: number) => `/api/sessions/${id}/next/${num}`,
  },
} as const satisfies Record<string, unknown>;
