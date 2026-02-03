import { create } from 'zustand';

export interface TestCase {
  input: string;
  expected_output: string;
  explanation?: string;
}

export interface FocusArea {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface Problem {
  id: string;
  title: string;
  description: string;
  focus_area_id: string;
  focus_area: FocusArea;
  sample_cases: TestCase[];
  hidden_cases: TestCase[];
  created_at: string;
}

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

interface InterviewStore {
  // Focus areas
  focusAreas: FocusArea[];
  selectedFocusAreas: string[];
  setFocusAreas: (areas: FocusArea[]) => void;
  toggleFocusArea: (slug: string) => void;

  // Current problem
  currentProblem: Problem | null;
  setCurrentProblem: (problem: Problem | null) => void;

  // Code editor
  code: string;
  setCode: (code: string) => void;

  // Execution results
  executionResults: ExecutionResponse | null;
  setExecutionResults: (results: ExecutionResponse | null) => void;

  // UI state
  isGenerating: boolean;
  isExecuting: boolean;
  setIsGenerating: (loading: boolean) => void;
  setIsExecuting: (loading: boolean) => void;
}

export const useInterviewStore = create<InterviewStore>((set) => ({
  // Focus areas
  focusAreas: [],
  selectedFocusAreas: [],
  setFocusAreas: (areas) => set({ focusAreas: areas }),
  toggleFocusArea: (slug) =>
    set((state) => ({
      selectedFocusAreas: state.selectedFocusAreas.includes(slug)
        ? state.selectedFocusAreas.filter((s) => s !== slug)
        : [...state.selectedFocusAreas, slug],
    })),

  // Current problem
  currentProblem: null,
  setCurrentProblem: (problem) => set({ currentProblem: problem }),

  // Code editor
  code: '#include <bits/stdc++.h>\nusing namespace std;\n\nvoid solve() {\n    int a,b;if (cin>>a>>b) cout<<(a+b)<<endl;\n}\n\nint main() {\n    ios_base::sync_with_stdio(false);\n    cin.tie(NULL);\n    int t;cin>>t; while(t--) solve();\n    return 0;\n}',
  setCode: (code) => set({ code }),

  // Execution results
  executionResults: null,
  setExecutionResults: (results) => set({ executionResults: results }),

  // UI state
  isGenerating: false,
  isExecuting: false,
  setIsGenerating: (loading) => set({ isGenerating: loading }),
  setIsExecuting: (loading) => set({ isExecuting: loading }),
}));
