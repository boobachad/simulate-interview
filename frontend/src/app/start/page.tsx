'use client';

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ClockIcon, Loader2Icon, PlayIcon, ListIcon } from "lucide-react";
import { InterviewSplitLayout } from "@/components/InterviewSplitLayout";
import { useInterviewStore } from "@/lib/store";
import { focusAreasApi, problemsApi, executionApi } from "@/lib/api";
import { apiClient } from "@/lib/api-client";
import { toast } from "sonner";
import { CodeEditor } from "@/components/CodeEditor";
import { TestCasesView } from "@/components/TestCasesView";
import { Navbar } from "@/components/Navbar";
import { FocusAreaSelector } from "@/components/FocusAreaSelector";
import type { FocusSelection, SessionCreateRequest } from "@/types/api";
import { isSingleTopicMode, isMultipleTopicsMode } from "@/types/api";


const CPP_BOILERPLATE = `#include <bits/stdc++.h>
using namespace std;

void solve() {
    // ============================================
    // START: Write your solution code here
    // ============================================
    
    int a, b;
    cin >> a >> b;
    cout << a + b << endl;
    
    // ============================================
    // END: Write your solution code above
    // ============================================
}

// DO NOT MODIFY BELOW THIS LINE
int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t;
    cin >> t;
    while(t--) solve();
    return 0;
}`;

const PYTHON_BOILERPLATE = `def solve():
    # ============================================
    # START: Write your solution code here
    # ============================================
    
    a, b = map(int, input().split())
    print(a + b)
    
    # ============================================
    # END: Write your solution code above
    # ============================================

# DO NOT MODIFY BELOW THIS LINE
if __name__ == "__main__":
    t = int(input())
    for _ in range(t):
        solve()`;

const JAVA_BOILERPLATE = `import java.util.*;
import java.io.*;

public class Main {
    public static void solve(Scanner sc) {
        // ============================================
        // START: Write your solution code here
        // ============================================
        
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(a + b);
        
        // ============================================
        // END: Write your solution code above
        // ============================================
    }
    
    // DO NOT MODIFY BELOW THIS LINE
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int t = sc.nextInt();
        while (t-- > 0) {
            solve(sc);
        }
        sc.close();
    }
}`;

const JAVASCRIPT_BOILERPLATE = `const readline = require('readline');
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

let lines = [];
rl.on('line', (line) => {
    lines.push(line);
}).on('close', () => {
    const t = parseInt(lines[0]);
    let idx = 1;
    
    for (let i = 0; i < t; i++) {
        // ============================================
        // START: Write your solution code here
        // ============================================
        
        const [a, b] = lines[idx++].split(' ').map(Number);
        console.log(a + b);
        
        // ============================================
        // END: Write your solution code above
        // ============================================
    }
});`;

const BOILERPLATES: Record<string, string> = {
    cpp: CPP_BOILERPLATE,
    python: PYTHON_BOILERPLATE,
    java: JAVA_BOILERPLATE,
    javascript: JAVASCRIPT_BOILERPLATE,
};

export default function StartPage() {
  const router = useRouter();
  const { focusAreas, setFocusAreas, selectedFocusAreas, toggleFocusArea, setCurrentProblem } = useInterviewStore();
  const [isGenerating, setIsGenerating] = useState(false);
  const [mode, setMode] = useState<"playground" | "session">("playground");
  const [focusSelection, setFocusSelection] = useState<FocusSelection>({ mode: "all" });
  const [problemCount, setProblemCount] = useState(5);


  // Editor & Execution State
  const [language, setLanguage] = useState("cpp");
  const [code, setCode] = useState(BOILERPLATES.cpp);
  const [isRunning, setIsRunning] = useState(false);
  const [executionResults, setExecutionResults] = useState<any[] | null>(null);
  const [executionError, setExecutionError] = useState<string | null>(null);
  const [customTestCases, setCustomTestCases] = useState<{ id: string; input: string }[]>([]);

  // Update code when language changes
  useEffect(() => {
    setCode(BOILERPLATES[language] || BOILERPLATES.cpp);
  }, [language]);

  useEffect(() => {
    loadFocusAreas();
  }, []);

  useEffect(() => {
    setExecutionResults(null);
  }, [customTestCases]);

  const loadFocusAreas = async () => {
    try {
      const areas = await focusAreasApi.getAll();
      setFocusAreas(areas);
    } catch (error) {
      console.error('Failed to load focus areas:', error);
    }
  };

  const handleGenerate = async () => {
    if (mode === "session") {
      await handleCreateSession();
    } else {
      await handleGenerateSingleProblem();
    }
  };

  const handleGenerateSingleProblem = async () => {
    setIsGenerating(true);

    try {
      const problem = await problemsApi.generate(selectedFocusAreas);
      setCurrentProblem(problem);
      router.push(`/problem/${problem.id}`);
    } catch (error) {
      console.error(error);
      toast.error("Failed to generate problem");
      setIsGenerating(false);
    }
  };

  const handleCreateSession = async () => {
    setIsGenerating(true);

    try {
      let request: SessionCreateRequest;

      if (focusSelection.mode === "all") {
        request = { focus_mode: "all", problem_count: problemCount };
      } else if (isSingleTopicMode(focusSelection)) {
        if (!focusSelection.topic) {
          toast.error("Please select a topic");
          setIsGenerating(false);
          return;
        }
        request = {
          focus_mode: "single",
          focus_topic: focusSelection.topic,
          problem_count: problemCount,
        };
      } else if (isMultipleTopicsMode(focusSelection)) {
        if (focusSelection.topics.length < 2 || focusSelection.topics.length > 10) {
          toast.error("Please select 2-10 topics");
          setIsGenerating(false);
          return;
        }
        request = {
          focus_mode: "multiple",
          focus_topics: focusSelection.topics,
          problem_count: problemCount,
        };
      } else {
        toast.error("Invalid focus selection");
        setIsGenerating(false);
        return;
      }

      const response = await apiClient.sessions.create(request);
      router.push(`/problem/${response.first_problem.id}?session=${response.session_id}`);
    } catch (error) {
      console.error(error);
      toast.error("Failed to create session");
      setIsGenerating(false);
    }
  };

  const executeCode = async (mode: "run" | "submit") => {
    setIsRunning(true);
    setExecutionResults(null);
    setExecutionError(null);

    try {
      const formattedCustomCases = customTestCases.map(c => ({
        input: c.input,
        expected_output: "",
      }));

      // Use "playground" ID for execution on Start Page to avoid running against mock problem sample cases
      const response = await executionApi.execute(code || "", "playground", language, formattedCustomCases, mode);

      if (response.success || response.results) {
        setExecutionResults(response.results);
        if (mode === "submit") {
          if (response.success) {
            toast.success(`ACCEPTED! Passed all ${response.total_cases} test cases.`);
          } else {
            toast.error(`Wrong Answer. Passed ${response.total_passed}/${response.total_cases} test cases.`);
          }
        } else {
          if (response.success) {
            toast.success(`Run Passed: ${response.total_passed}/${response.total_cases} cases.`);
          } else {
            toast.error(`Run Failed: ${response.total_passed}/${response.total_cases} cases.`);
          }
        }
      } else {
        setExecutionError("Execution failed to return results");
      }
    } catch (error: any) {
      const msg = error.response?.data?.error || "Execution failed";
      setExecutionError(msg);
    } finally {
      setIsRunning(false);
    }
  };

  const handleRun = () => executeCode("run");
  const handleSubmit = () => executeCode("submit");

  // --- Left Content: Sidebar ---
  const LeftContent = (
    <div className="flex flex-col gap-4 min-h-0 h-full">
      {/* Mode Toggle */}
      <Tabs value={mode} onValueChange={(v) => setMode(v as "playground" | "session")} className="flex-shrink-0">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="playground" className="flex items-center gap-2">
            <PlayIcon className="h-4 w-4" />
            Playground
          </TabsTrigger>
          <TabsTrigger value="session" className="flex items-center gap-2">
            <ListIcon className="h-4 w-4" />
            Session
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Status Indicator */}
      <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex items-center gap-2 flex-shrink-0">
        {isGenerating ? <Loader2Icon className="h-4 w-4 animate-spin" /> : <ClockIcon className="h-4 w-4 text-muted-foreground" />}
        <span className="text-sm font-medium">
          {isGenerating ? (mode === "session" ? "Creating session..." : "Generating problem...") : "Ready to generate"}
        </span>
      </div>

      {/* Focus Area Selection */}
      {mode === "session" ? (
        <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex-1 min-h-0 overflow-y-auto">
          <FocusAreaSelector
            selection={focusSelection}
            onSelectionChange={setFocusSelection}
          />
          
          {/* Problem Count Selector */}
          <div className="mt-4 pt-4 border-t">
            <label htmlFor="problem-count" className="text-sm font-semibold mb-2 block">
              Problem Count
            </label>
            <select
              id="problem-count"
              value={problemCount}
              onChange={(e) => setProblemCount(Number(e.target.value))}
              className="w-full px-3 py-2 rounded-md border bg-background text-sm"
            >
              {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((n) => (
                <option key={n} value={n}>
                  {n} {n === 1 ? "problem" : "problems"}
                </option>
              ))}
            </select>
          </div>
        </div>
      ) : (
        <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex flex-col gap-4 flex-shrink-0">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-sm">Focus Areas</h3>
            <Badge variant="secondary" className="text-xs">
              {selectedFocusAreas.length === 0 ? "Random" : `${selectedFocusAreas.length} selected`}
            </Badge>
          </div>

          <div className="flex flex-wrap gap-2 content-start">
            <div
              className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${selectedFocusAreas.length === 0 ? "bg-primary text-primary-foreground border-primary" : "bg-secondary text-secondary-foreground hover:bg-secondary/80"}`}
              onClick={() => setFocusAreas(focusAreas)}
            >
              All (Random)
            </div>
            {focusAreas.map(area => (
              <div
                key={area.id}
                onClick={() => toggleFocusArea(area.slug)}
                className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${selectedFocusAreas.includes(area.slug)
                  ? "bg-primary text-primary-foreground border-primary"
                  : "bg-background text-foreground border-input hover:bg-accent hover:text-accent-foreground"
                  }`}
              >
                {area.name}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Generate Button */}
      <Button
        className="w-full h-10 font-semibold flex-shrink-0"
        onClick={handleGenerate}
        disabled={isGenerating}
      >
        {isGenerating 
          ? (mode === "session" ? "Creating Session..." : "Generating Problem...") 
          : (mode === "session" ? "Start Session" : "Generate Problem")}
      </Button>

      {mode === "playground" && (
        <>
          <div className="flex-1 min-h-0"></div>

          {/* Difficulty Buttons */}
          <div className="flex gap-2 flex-shrink-0 mt-auto">
            <Button variant="outline" className="flex-1" disabled>Make Easier</Button>
            <Button variant="outline" className="flex-1" disabled>Make Harder</Button>
          </div>
        </>
      )}
    </div>
  );

  // --- Right Content: Editor & Tests ---
  const RightTop = (
    <CodeEditor
      language={language}
      setLanguage={setLanguage}
      code={code || ""}
      setCode={setCode}
      onRun={handleRun}
      onSubmit={handleSubmit}
      isRunning={isRunning}
      readOnly={false}
    />
  );

  const RightBottom = (
    <TestCasesView
      testCases={[]}
      results={executionResults}
      customTestCases={customTestCases}
      setCustomTestCases={setCustomTestCases}
      error={executionError}
    />
  );

  return (
    <div className="h-screen w-screen flex flex-col overflow-hidden bg-background">
      <Navbar breadcrumbItems={[{ label: 'home', href: '/' }, { label: 'start' }]} />
      <div className="flex-1 min-h-0">
        <InterviewSplitLayout
          leftContent={LeftContent}
          rightTopContent={RightTop}
          rightBottomContent={RightBottom}
        />
      </div>
    </div>
  );
}
