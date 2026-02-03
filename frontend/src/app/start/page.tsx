'use client';

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ClockIcon, Loader2Icon } from "lucide-react";
import { InterviewSplitLayout } from "@/components/InterviewSplitLayout";
import { useInterviewStore } from "@/lib/store";
import { focusAreasApi, problemsApi, executionApi } from "@/lib/api";
import { toast } from "sonner";
import { CodeEditor } from "@/components/CodeEditor";
import { TestCasesView } from "@/components/TestCasesView";
import { Navbar } from "@/components/Navbar";


const CPP_BOILERPLATE = `#include <bits/stdc++.h>
using namespace std;

void solve() {
    int a,b;if (cin>>a>>b) cout<<(a+b)<<endl;
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t;cin>>t; while(t--) solve();
    return 0;
}`;

export default function StartPage() {
  const router = useRouter();
  const { focusAreas, setFocusAreas, selectedFocusAreas, toggleFocusArea, setCurrentProblem } = useInterviewStore();
  const [isGenerating, setIsGenerating] = useState(false);


  // Editor & Execution State
  const [language, setLanguage] = useState("cpp");
  const [code, setCode] = useState(CPP_BOILERPLATE);
  const [isRunning, setIsRunning] = useState(false);
  const [executionResults, setExecutionResults] = useState<any[] | null>(null);
  const [customTestCases, setCustomTestCases] = useState<{ id: string; input: string }[]>([]);

  useEffect(() => {
    loadFocusAreas();
  }, []);

  const loadFocusAreas = async () => {
    try {
      const areas = await focusAreasApi.getAll();
      setFocusAreas(areas);
    } catch (error) {
      console.error('Failed to load focus areas:', error);
    }
  };

  const handleGenerate = async () => {
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

  const executeCode = async (mode: "run" | "submit") => {
    setIsRunning(true);
    setExecutionResults(null);

    try {
      const formattedCustomCases = customTestCases.map(c => ({
        input: c.input,
        expected_output: "",
      }));

      // Use "playground" ID for execution on Start Page to avoid running against mock problem sample cases
      const response = await executionApi.execute(code, "playground", formattedCustomCases, mode);

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
        toast.error("Execution failed to return results");
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Execution failed");
    } finally {
      setIsRunning(false);
    }
  };

  const handleRun = () => executeCode("run");
  const handleSubmit = () => executeCode("submit");

  // --- Left Content: Sidebar ---
  const LeftContent = (
    <div className="flex flex-col gap-4 min-h-0 h-full">
      {/* 1. Status Indicator */}
      <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex items-center gap-2 flex-shrink-0">
        {isGenerating ? <Loader2Icon className="h-4 w-4 animate-spin" /> : <ClockIcon className="h-4 w-4 text-muted-foreground" />}
        <span className="text-sm font-medium">
          {isGenerating ? "Generating problem..." : "Ready to generate"}
        </span>
      </div>

      {/* 2. Focus Area Selector */}
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

      {/* 3. Generate Button */}
      <Button
        className="w-full h-10 font-semibold flex-shrink-0"
        onClick={handleGenerate}
        disabled={isGenerating}
      >
        {isGenerating ? "Generating New Problem..." : "Generate New Problem"}
      </Button>

      <div className="flex-1 min-h-0"></div>

      {/* 5. Difficulty Buttons */}
      <div className="flex gap-2 flex-shrink-0 mt-auto">
        <Button variant="outline" className="flex-1" disabled>Make Easier</Button>
        <Button variant="outline" className="flex-1" disabled>Make Harder</Button>
      </div>
    </div>
  );

  // --- Right Content: Editor & Tests ---
  const RightTop = (
    <CodeEditor
      language={language}
      setLanguage={setLanguage}
      code={code}
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
