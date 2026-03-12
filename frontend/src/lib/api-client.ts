import type {
  LoginRequest,
  LoginResponse,
  ProfileRequest,
  ProfileData,
  CombinedStats,
  FocusArea,
  SessionCreateRequest,
  SessionCreateResponse,
  SessionData,
  NextProblemResponse,
  SessionID,
} from "@/types/api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Error types for specific HTTP status codes
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

// Retry configuration
const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 1000;
const RETRYABLE_STATUS_CODES = [429, 503, 504];

// Sleep utility for retry delays
const sleep = (ms: number): Promise<void> =>
  new Promise((resolve) => setTimeout(resolve, ms));

// Check if error is a network error
function isNetworkError(error: unknown): boolean {
  return error instanceof TypeError || (error as Error).name === "TypeError";
}

// Get auth token from localStorage
function getAuthToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("auth_token");
}

// Set auth token in localStorage
export function setAuthToken(token: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem("auth_token", token);
}

// Clear auth token from localStorage
export function clearAuthToken(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem("auth_token");
}

// Fetch wrapper with retry logic and error handling
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
    const response = await fetch(`${API_BASE_URL}${url}`, {
      ...options,
      headers,
    });

    // Handle specific error status codes
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

    // Parse response
    const contentType = response.headers.get("content-type");
    const contentLength = response.headers.get("content-length");

    // Handle empty or non-JSON responses
    if (
      response.status === 204 ||
      contentLength === "0" ||
      !contentType?.includes("application/json")
    ) {
      return {} as T;
    }

    // Parse JSON with error handling
    try {
      return (await response.json()) as T;
    } catch (parseError) {
      console.error("Failed to parse response JSON:", parseError);
      return {} as T;
    }
  } catch (error) {
    // Retry on transient failures (HTTP errors or network errors)
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
export const apiClient = {
  // Authentication
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

  // Profile
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

  // Stats
  stats: {
    get: async (): Promise<CombinedStats> => {
      return fetchWithRetry<CombinedStats>("/api/stats");
    },
  },

  // Focus Areas
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

  // Sessions
  sessions: {
    create: async (
      data: SessionCreateRequest
    ): Promise<SessionCreateResponse> => {
      return fetchWithRetry<SessionCreateResponse>("/api/sessions", {
        method: "POST",
        body: JSON.stringify(data),
      });
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
  },
};
