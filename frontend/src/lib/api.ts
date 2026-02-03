import axios from 'axios';
import { FocusArea, Problem, ExecutionResponse } from './store';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const focusAreasApi = {
  getAll: async (): Promise<FocusArea[]> => {
    const response = await api.get('/focus-areas');
    return response.data;
  },
};

export const problemsApi = {
  generate: async (focusAreas: string[]): Promise<Problem> => {
    const response = await api.post('/problems/generate', {
      focus_areas: focusAreas,
    });
    return response.data;
  },

  getById: async (id: string): Promise<Problem> => {
    const response = await api.get(`/problems/${id}`);
    return response.data;
  },

  getAll: async (focusArea?: string): Promise<Problem[]> => {
    const params = focusArea ? { focus_area: focusArea } : {};
    const response = await api.get('/problems', { params });
    return response.data;
  },
};

export const executionApi = {
  execute: async (code: string, problemId: string, customCases?: any[], mode: "run" | "submit" = "run"): Promise<ExecutionResponse> => {
    const response = await api.post('/execute', {
      code,
      problem_id: problemId,
      custom_cases: customCases || [],
      mode,
    });
    return response.data;
  },
};

export interface SkillLevel {
  level: string;
  problem_count: number;
  tag: string;
}

export interface LeetCodeStats {
  username: string;
  ranking: number;
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
  rank: string;
  max_rank: string;
  problems_solved: number;
  tags: Record<string, number>;
}

export interface UserProfile {
  name: string;
  leetcode_stats?: LeetCodeStats;
  codeforces_stats?: CodeforcesStats;
  suggested_areas: string[];
}

export const statsApi = {
  getUserStats: async (
    name: string,
    leetcodeUsername?: string,
    codeforcesUsername?: string
  ): Promise<UserProfile> => {
    const response = await api.post('/stats', {
      name,
      leetcode_username: leetcodeUsername || '',
      codeforces_username: codeforcesUsername || '',
    });
    return response.data;
  },
};
