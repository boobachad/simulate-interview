import { create } from "zustand";
import type { FocusArea, Problem, ExecutionResponse, TestCase } from "./api";

interface InterviewStore {
  // Focus areas (dynamic platform topics)
  focusAreas: FocusArea[];
  selectedFocusAreas: string[];
  setFocusAreas: (areas: FocusArea[]) => void;
  toggleFocusArea: (topic: string) => void;
  clearSelectedFocusAreas: () => void;

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
  toggleFocusArea: (topic) =>
    set((state) => ({
      selectedFocusAreas: state.selectedFocusAreas.includes(topic)
        ? state.selectedFocusAreas.filter((t) => t !== topic)
        : [...state.selectedFocusAreas, topic],
    })),
  clearSelectedFocusAreas: () => set({ selectedFocusAreas: [] }),

  // Current problem
  currentProblem: null,
  setCurrentProblem: (problem) => set({ currentProblem: problem }),

  // Code editor
  code: "#include <bits/stdc++.h>\nusing namespace std;\n\nvoid solve() {\n    int a,b;if (cin>>a>>b) cout<<(a+b)<<endl;\n}\n\nint main() {\n    ios_base::sync_with_stdio(false);\n    cin.tie(NULL);\n    int t;cin>>t; while(t--) solve();\n    return 0;\n}",
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
