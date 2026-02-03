'use client';

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2Icon, PlayIcon, SendIcon, ClockIcon, LockIcon, UnlockIcon } from "lucide-react";
import { InterviewSplitLayout } from "@/components/InterviewSplitLayout";
import { useInterviewStore } from "@/lib/store";
import { problemsApi, executionApi } from "@/lib/api"; // Added executionApi
import { toast } from "sonner";
import { CodeEditor } from "@/components/CodeEditor";
import { Skeleton } from "@/components/ui/skeleton";
import ReactMarkdown from "react-markdown";
import { TestCasesView } from "@/components/TestCasesView";
import { InterviewTimer } from "@/components/InterviewTimer";
import { Navbar } from "@/components/Navbar";

const CPP_BOILERPLATE = `#include <bits/stdc++.h>
using namespace std;

void solve() {
    // Read your input here
    int a,b; if (cin>>a>>b) cout<<(a+b)<<endl;
}

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t;cin>>t; while(t--) solve();
    return 0;
}`;

export default function ProblemPage() {
  const params = useParams();
  const router = useRouter();
  const problemId = params.id as string;
  const { currentProblem, setCurrentProblem } = useInterviewStore();

  const [language, setLanguage] = useState("cpp");
  const [code, setCode] = useState(CPP_BOILERPLATE);
  const [isRunning, setIsRunning] = useState(false);

  // Custom Test Cases State lifted from TestCasesView
  const [customTestCases, setCustomTestCases] = useState<{ id: string; input: string }[]>([]);

  // Store execution results
  const [executionResults, setExecutionResults] = useState<any[] | null>(null);

  useEffect(() => {
    if (!currentProblem || currentProblem.id !== problemId) {
      loadProblem();
    }
  }, [problemId]);

  const loadProblem = async () => {
    try {
      const problem = await problemsApi.getById(problemId);
      setCurrentProblem(problem);
    } catch (error) {
      toast.error("Failed to load problem");
    }
  };

  const handleRun = async () => {
    await executeCode("run");
  };

  const handleSubmit = async () => {
    await executeCode("submit");
  };

  const executeCode = async (mode: "run" | "submit") => {
    setIsRunning(true);
    setExecutionResults(null);

    try {
      // Format custom cases for the API
      const formattedCustomCases = customTestCases.map(c => ({
        input: c.input,
        expected_output: "",
      }));

      const response = await executionApi.execute(code, problemId, formattedCustomCases, mode);

      if (response.success || response.results) {
        setExecutionResults(response.results);

        // Different success message for Submit vs Run
        if (mode === "submit") {
          if (response.success) {
            toast.success(`ACCEPTED! Passed all ${response.total_cases} test cases.`);
          } else {
            // Wrong Answer: Apply Penalty
            const penaltySeconds = Number(process.env.NEXT_PUBLIC_WRONG_SUBMISSION_PENALTY_MINUTES || 2) * 60;
            setTimeLeft(prev => Math.max(0, prev - penaltySeconds));
            toast.error(`Wrong Answer. Passed ${response.total_passed}/${response.total_cases} test cases.`);
            toast.warning(`${penaltySeconds / 60} minutes penalty applied`);
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

  // --- Timer & Hints State ---
  const durationMinutes = Number(process.env.NEXT_PUBLIC_INTERVIEW_DURATION_MINUTES) || 30;
  const solutionPenaltyMinutes = Number(process.env.NEXT_PUBLIC_SOLUTION_PENALTY_MINUTES) || 5;
  const wrongSubmissionPenaltyMinutes = Number(process.env.NEXT_PUBLIC_WRONG_SUBMISSION_PENALTY_MINUTES) || 2;

  const [timeLeft, setTimeLeft] = useState(durationMinutes * 60);
  const [isHintsUnlocked, setIsHintsUnlocked] = useState(false);

  // Auto-unlock hints when time runs out
  useEffect(() => {
    if (timeLeft <= 0 && !isHintsUnlocked) {
      setIsHintsUnlocked(true);
      toast.info("Time's up! Hints have been unlocked.");
    }
  }, [timeLeft, isHintsUnlocked]);

  // Handle Unlock Hint
  const handleUnlockHint = () => {
    const penalty = solutionPenaltyMinutes * 60;

    if (timeLeft <= penalty) {
      setTimeLeft(0);
      setIsHintsUnlocked(true);
      toast.warning("Time exhausted to unlock hints!");
    } else {
      setTimeLeft(prev => prev - penalty);
      setIsHintsUnlocked(true);
      toast.success(`Hints unlocked (-${solutionPenaltyMinutes}:00 penalty applied)`);
    }
  };

  // Split description and hints
  let descriptionDisplay = "";
  let hintsDisplay = "";

  if (currentProblem?.description) {
    const parts = currentProblem.description.split("## Solution Hints");
    descriptionDisplay = parts[0];
    if (parts.length > 1) {
      hintsDisplay = parts[1];
    }
  }

  // --- Left Content: Problem Details ---
  const LeftContent = !currentProblem ? (
    <div className="space-y-4">
      <Skeleton className="h-8 w-3/4" />
      <Skeleton className="h-32 w-full" />
    </div>
  ) : (
    <div className="flex flex-col gap-4 min-h-0 h-full">
      {/* 1. Timer (Replaces static status) */}
      <InterviewTimer
        timeLeft={timeLeft}
        setTimeLeft={setTimeLeft}
        totalTime={durationMinutes * 60}
      />

      {/* 2. Focus Areas Summary */}
      <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex flex-col gap-2 flex-shrink-0">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold text-sm">Focus Areas</h3>
          <Badge variant="secondary" className="text-xs">
            {currentProblem.focus_area?.name || "General"}
          </Badge>
        </div>
      </div>

      {/* 3. Generate Button */}
      <Button
        className="w-full h-10 font-semibold flex-shrink-0"
        onClick={() => router.push('/start')}
      >
        Generate New Problem
      </Button>

      {/* 4. Generated Content (Scrollable) */}
      <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 space-y-6 flex-1 overflow-auto min-h-0 relative">
        <div className="space-y-2">
          <h2 className="font-bold text-lg font-comic">{currentProblem.title}</h2>
        </div>

        {/* Main Description */}
        <div className="prose dark:prose-invert text-sm max-w-none">
          <ReactMarkdown>{descriptionDisplay}</ReactMarkdown>
        </div>

        {/* Hints Section */}
        {hintsDisplay && (
          <div className="mt-8 pt-6 border-t border-border">
            <h3 className="font-semibold text-sm text-muted-foreground mb-3 flex items-center gap-2 uppercase tracking-wider">
              {isHintsUnlocked ? <UnlockIcon className="w-4 h-4 text-success" /> : <LockIcon className="w-4 h-4 text-warning" />}
              Solution Hints
            </h3>

            {isHintsUnlocked ? (
              <div className="text-sm text-foreground/90 leading-relaxed animate-in fade-in slide-in-from-bottom-2">
                <ReactMarkdown>{hintsDisplay}</ReactMarkdown>
              </div>
            ) : (
              <div className="group relative overflow-hidden rounded-md border bg-muted/30 p-8 text-center transition-all hover:bg-muted/50">
                {/* Blur overlay effect */}
                <div className="absolute inset-0 flex items-center justify-center backdrop-blur-[1px] bg-background/50 pointer-events-none" />

                <div className="relative z-10 flex flex-col items-center gap-4">
                  <div className="p-3 rounded-full bg-background border shadow-sm">
                    <LockIcon className="w-5 h-5 text-muted-foreground" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-foreground">
                      Hints Locked
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Unlock cost: <span className="font-mono text-destructive">-{solutionPenaltyMinutes}:00</span>
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleUnlockHint}
                    className="mt-2 transition-all hover:border-destructive hover:text-destructive active:scale-95"
                  >
                    Unlock Hints
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 5. Difficulty Buttons */}
      <div className="flex gap-2 flex-shrink-0 mt-auto pt-2">
        <Button variant="outline" className="flex-1">Make Easier</Button>
        <Button variant="outline" className="flex-1">Make Harder</Button>
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
      isRunning={isRunning}
      onSubmit={handleSubmit}
    />
  );

  const RightBottom = (
    <TestCasesView
      testCases={currentProblem?.sample_cases || []}
      results={executionResults}
      customTestCases={customTestCases}
      setCustomTestCases={setCustomTestCases}
    />
  );

  return (
    <div className="h-screen w-screen flex flex-col overflow-hidden bg-background">
      <Navbar breadcrumbItems={[
        { label: 'home', href: '/' },
        { label: 'problem', href: '/problem' },
        { label: currentProblem?.title || 'loading...' }
      ]} />
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
