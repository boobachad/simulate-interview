// Branded types for domain IDs
type Brand<T, B extends string> = T & { readonly __brand: B };

export type UserID = Brand<string, "UserID">;
export type SessionID = Brand<string, "SessionID">;
export type ProblemID = Brand<string, "ProblemID">;

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
  default_problem_count?: number;
}

export interface ProfileData {
  leetcode_username: string;
  codeforces_username: string;
  last_synced: string | null;
  problem_count: number;
  default_problem_count: number;
}

// Stats
export interface SkillLevel {
  level: string;
  problem_count: number;
  tag: string;
}

export interface LeetCodeStats {
  username: string;
  total_solved: number;
  easy_solved: number;
  medium_solved: number;
  hard_solved: number;
  skills: Record<string, SkillLevel>;
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

// Focus Areas (Dynamic Platform Topics)
export interface FocusArea {
  platform: string;
  topic: string;
  problem_count: number;
  user_solved?: number;
}

// Focus Selection
export type FocusSelection =
  | { mode: "all" }
  | { mode: "single"; topic: string }
  | { mode: "multiple"; topics: string[] };

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
  rating: number;
  focus_area: string;
  sample_cases: TestCase[];
  created_at: string;
}

// Execution
export interface ExecutionResult {
  case_number: number;
  input: string;
  expected_output: string;
  actual_output: string;
  passed: boolean;
  error?: string;
}

export interface ExecutionResponse {
  success: boolean;
  results: ExecutionResult[];
  total_passed: number;
  total_cases: number;
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

export type SessionProblemStatus = "generating" | "ready" | "failed";

export interface SessionProblem {
  problem_number: number;
  status: SessionProblemStatus;
  problem?: Problem;
  error_message?: string;
}

export function isReady(
  problem: SessionProblem
): problem is SessionProblem & { status: "ready"; problem: Problem } {
  return problem.status === "ready" && problem.problem !== undefined;
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

export interface ActiveSessionSummary {
  id: SessionID;
  problem_count: number;
  current_problem_number: number;
  ready_problems: number;
  focus_mode: "all" | "single" | "multiple";
  first_ready_problem_id?: string;
  created_at: string;
}

// API Error Classes
export class APIError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    message: string
  ) {
    super(message);
    this.name = "APIError";
  }
}

export class UnauthorizedError extends APIError {
  constructor(message = "Unauthorized") {
    super(401, "Unauthorized", message);
    this.name = "UnauthorizedError";
  }
}

export class RateLimitError extends APIError {
  constructor(message = "Rate limit exceeded") {
    super(429, "Too Many Requests", message);
    this.name = "RateLimitError";
  }
}

export class ServiceUnavailableError extends APIError {
  constructor(message = "Service unavailable") {
    super(503, "Service Unavailable", message);
    this.name = "ServiceUnavailableError";
  }
}

export class GatewayTimeoutError extends APIError {
  constructor(message = "Gateway timeout") {
    super(504, "Gateway Timeout", message);
    this.name = "GatewayTimeoutError";
  }
}

// Configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 1000;
const RETRYABLE_STATUS_CODES = [429, 503, 504];

const sleep = (ms: number): Promise<void> =>
  new Promise((resolve) => setTimeout(resolve, ms));

function isNetworkError(error: unknown): boolean {
  return error instanceof TypeError || (error as Error).name === "TypeError";
}

function getAuthToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("auth_token");
}

export function setAuthToken(token: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem("auth_token", token);
  document.cookie = `auth_token=${token}; path=/; max-age=${24 * 60 * 60}; SameSite=Lax`;
}

export function clearAuthToken(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem("auth_token");
  document.cookie = "auth_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
}

async function fetchWithRetry<T>(
  url: string,
  options: RequestInit = {},
  retryCount = 0
): Promise<T> {
  const token = getAuthToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  try {
    const fetchOptions: RequestInit = {
      ...options,
      headers,
    };

    if (options.signal) {
      fetchOptions.signal = options.signal;
    }

    const response = await fetch(`${API_BASE_URL}${url}`, fetchOptions);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage =
        (errorData as { error?: string }).error || response.statusText;

      switch (response.status) {
        case 401:
          clearAuthToken();
          throw new UnauthorizedError(errorMessage);
        case 429:
          throw new RateLimitError(errorMessage);
        case 503:
          throw new ServiceUnavailableError(errorMessage);
        case 504:
          throw new GatewayTimeoutError(errorMessage);
        default:
          throw new APIError(response.status, response.statusText, errorMessage);
      }
    }

    const contentType = response.headers.get("content-type");
    const contentLength = response.headers.get("content-length");

    if (
      response.status === 204 ||
      contentLength === "0" ||
      !contentType?.includes("application/json")
    ) {
      return {} as T;
    }

    try {
      return (await response.json()) as T;
    } catch (parseError) {
      // Check if response body is actually empty
      const clonedResponse = response.clone();
      const text = await clonedResponse.text();
      
      if (text.trim().length === 0) {
        // Truly empty response, return empty object
        return {} as T;
      }
      
      // Non-empty body that failed to parse - this is a real error
      console.error("Failed to parse non-empty response JSON:", parseError);
      console.error("Response body:", text);
      throw new Error(`JSON parse error: ${parseError instanceof Error ? parseError.message : String(parseError)}`);
    }
  } catch (error) {
    const shouldRetry =
      retryCount < MAX_RETRIES &&
      ((error instanceof APIError &&
        RETRYABLE_STATUS_CODES.includes(error.status)) ||
        isNetworkError(error));

    if (shouldRetry) {
      const delay = RETRY_DELAY_MS * Math.pow(2, retryCount);
      await sleep(delay);
      return fetchWithRetry<T>(url, options, retryCount + 1);
    }

    throw error;
  }
}

// API Client
export const api = {
  auth: {
    login: async (data: LoginRequest): Promise<LoginResponse> => {
      return fetchWithRetry<LoginResponse>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify(data),
      });
    },

    logout: async (): Promise<{ success: boolean }> => {
      return fetchWithRetry<{ success: boolean }>("/api/auth/logout", {
        method: "POST",
      });
    },
  },

  profile: {
    setup: async (
      data: ProfileRequest
    ): Promise<{ success: boolean; sync_status: string }> => {
      return fetchWithRetry<{ success: boolean; sync_status: string }>(
        "/api/profile/setup",
        {
          method: "POST",
          body: JSON.stringify(data),
        }
      );
    },

    get: async (): Promise<ProfileData> => {
      return fetchWithRetry<ProfileData>("/api/profile");
    },

    update: async (data: ProfileRequest): Promise<{ success: boolean }> => {
      return fetchWithRetry<{ success: boolean }>("/api/profile", {
        method: "PUT",
        body: JSON.stringify(data),
      });
    },

    sync: async (): Promise<{ success: boolean; synced_at: string }> => {
      return fetchWithRetry<{ success: boolean; synced_at: string }>(
        "/api/profile/sync",
        {
          method: "POST",
        }
      );
    },
  },

  stats: {
    get: async (signal?: AbortSignal): Promise<CombinedStats> => {
      const options: RequestInit = {};
      if (signal) {
        options.signal = signal;
      }
      return fetchWithRetry<CombinedStats>("/api/stats", options);
    },
  },

  focusAreas: {
    list: async (params?: {
      page?: number;
      page_size?: number;
    }): Promise<FocusArea[]> => {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.set("page", params.page.toString());
      if (params?.page_size)
        queryParams.set("page_size", params.page_size.toString());

      const query = queryParams.toString();
      const url = query ? `/api/focus-areas?${query}` : "/api/focus-areas";

      return fetchWithRetry<FocusArea[]>(url);
    },
  },

  problems: {
    generate: async (focusAreas: string[], targetRating?: number): Promise<Problem> => {
      const body: { focus_areas: string[]; target_rating?: number } = { focus_areas: focusAreas };
      if (targetRating !== undefined) {
        body.target_rating = targetRating;
      }
      return fetchWithRetry<Problem>("/api/problems/generate", {
        method: "POST",
        body: JSON.stringify(body),
      });
    },

    getById: async (id: string): Promise<Problem> => {
      return fetchWithRetry<Problem>(`/api/problems/${id}`);
    },

    getAll: async (focusArea?: string): Promise<Problem[]> => {
      const params = focusArea ? `?focus_area=${focusArea}` : "";
      return fetchWithRetry<Problem[]>(`/api/problems${params}`);
    },

    getSession: async (id: string): Promise<Problem> => {
      return fetchWithRetry<Problem>(`/api/problems/${id}/session`);
    },
  },

  execution: {
    execute: async (
      code: string,
      problemID: string,
      language: string = "cpp",
      customCases?: TestCase[],
      mode: "run" | "submit" = "run"
    ): Promise<ExecutionResponse> => {
      return fetchWithRetry<ExecutionResponse>("/api/execute", {
        method: "POST",
        body: JSON.stringify({
          code,
          problem_id: problemID,
          language,
          custom_cases: customCases || [],
          mode,
        }),
      });
    },
  },

  sessions: {
    create: async (
      data: SessionCreateRequest
    ): Promise<SessionCreateResponse> => {
      return fetchWithRetry<SessionCreateResponse>("/api/sessions", {
        method: "POST",
        body: JSON.stringify(data),
      });
    },

    list: async (): Promise<{ sessions: ActiveSessionSummary[] }> => {
      return fetchWithRetry<{ sessions: ActiveSessionSummary[] }>("/api/sessions");
    },

    get: async (sessionId: SessionID): Promise<SessionData> => {
      return fetchWithRetry<SessionData>(`/api/sessions/${sessionId}`);
    },

    getNext: async (
      sessionId: SessionID,
      currentNumber: number
    ): Promise<NextProblemResponse> => {
      return fetchWithRetry<NextProblemResponse>(
        `/api/sessions/${sessionId}/next/${currentNumber}`
      );
    },

    complete: async (sessionId: SessionID): Promise<{ success: boolean }> => {
      return fetchWithRetry<{ success: boolean }>(
        `/api/sessions/${sessionId}/complete`,
        {
          method: "POST",
        }
      );
    },
  },
};
