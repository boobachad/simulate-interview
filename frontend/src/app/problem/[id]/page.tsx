"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  LockIcon,
  UnlockIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ClockIcon,
} from "lucide-react";
import { InterviewSplitLayout } from "@/components/InterviewSplitLayout";
import { useInterviewStore } from "@/lib/store";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { CodeEditor } from "@/components/CodeEditor";
import { Skeleton } from "@/components/ui/skeleton";
import ReactMarkdown from "react-markdown";
import { TestCasesView } from "@/components/TestCasesView";
import { InterviewTimer } from "@/components/InterviewTimer";
import { Navbar } from "@/components/Navbar";
import { useSessionTimer } from "@/hooks/useSessionTimer";
import { useProblemTimer } from "@/hooks/useProblemTimer";
import { useCodePersistence } from "@/hooks/useCodePersistence";
import { ProblemTimer } from "@/components/ProblemTimer";
import { getRatingColor, adjustRating } from "@/lib/rating-utils";
import type { SessionData, SessionID } from "@/lib/api";

export default function ProblemPage() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const problemId = params.id as string;
  const sessionIdParam = searchParams.get("session");
  const { currentProblem, setCurrentProblem } = useInterviewStore();

  // Initialize isSessionMode immediately based on URL
  const [isSessionMode, setIsSessionMode] = useState(!!sessionIdParam);
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [currentProblemIndex, setCurrentProblemIndex] = useState(0);
  const [timerExtensionCount, setTimerExtensionCount] = useState(0);

  // Session timer hook (only for session mode)
  const {
    timeLeft: sessionTimeLeft,
    totalTime: sessionTotalTime,
    setTimeLeft: setSessionTimeLeft,
  } = useSessionTimer(sessionIdParam, sessionData);

  // Standalone timer for playground mode (30 minutes)
  const [playgroundTimeLeft, setPlaygroundTimeLeft] = useState(30 * 60);

  // Use appropriate timer based on mode
  const timeLeft = isSessionMode ? sessionTimeLeft : playgroundTimeLeft;
  const totalTime = isSessionMode ? sessionTotalTime : 30 * 60;
  const setTimeLeft = isSessionMode
    ? setSessionTimeLeft
    : setPlaygroundTimeLeft;

  // Problem timer hook (count-up for tracking)
  const { problemTime } = useProblemTimer(problemId, sessionIdParam);

  const [language, setLanguage] = useState("cpp");
  const [isRunning, setIsRunning] = useState(false);

  // Code persistence hook
  const { code, saveCode, cleanupSessionCode } = useCodePersistence(
    problemId,
    language,
    sessionIdParam,
  );

  const [customTestCases, setCustomTestCases] = useState<
    { id: string; input: string }[]
  >([]);
  const [executionResults, setExecutionResults] = useState<any[] | null>(null);
  const [executionError, setExecutionError] = useState<string | null>(null);

  // Standalone timer countdown for playground mode
  useEffect(() => {
    if (isSessionMode) return; // Only run in playground mode

    const interval = setInterval(() => {
      setPlaygroundTimeLeft((prev) => Math.max(0, prev - 1));
    }, 1000);

    return () => clearInterval(interval);
  }, [isSessionMode]);

  // Auto-complete session when timer expires (session mode only)
  useEffect(() => {
    // Only trigger when timer reaches exactly 0
    if (!isSessionMode || !sessionIdParam || !sessionData || timeLeft !== 0)
      return;

    const MAX_EXTENSIONS = 5;
    const readyCount = sessionData.problems.filter(
      (p) => p.status === "ready",
    ).length;

    if (readyCount < sessionData.problem_count) {
      // Timer expired but problems still generating
      if (timerExtensionCount >= MAX_EXTENSIONS) {
        toast.error(
          `Maximum timer extensions reached. Problems still generating (${readyCount}/${sessionData.problem_count}).`,
        );
        cleanupSessionCode();
        router.push("/start");
        return;
      }

      // Extend timer by 2 minutes
      const extensionSeconds = 2 * 60;
      setSessionTimeLeft((prev) => prev + extensionSeconds);
      setTimerExtensionCount((prev) => prev + 1);
      toast.warning(
        `Problems still generating. Timer extended by 2 minutes (${timerExtensionCount + 1}/${MAX_EXTENSIONS}).`,
      );
      return;
    }

    // All problems ready, auto-complete session
    const autoComplete = async () => {
      try {
        await api.sessions.complete(sessionIdParam as SessionID);
        cleanupSessionCode();
        setTimerExtensionCount(0);
        toast.info("Time expired! Session completed automatically.");
        router.push("/start");
      } catch (error) {
        console.error("Failed to auto-complete session:", error);
        toast.error("Failed to complete session");
      }
    };

    autoComplete();
  }, [
    isSessionMode,
    sessionIdParam,
    sessionData,
    timeLeft,
    timerExtensionCount,
    cleanupSessionCode,
    router,
    setSessionTimeLeft,
  ]);

  // Session detection and validation (runs once on mount or when session/problem ID changes)
  useEffect(() => {
    const detectSession = async () => {
      if (sessionIdParam) {
        setIsSessionMode(true);
        try {
          const session = await api.sessions.get(sessionIdParam as SessionID);
          setSessionData(session);

          const problemIdx = session.problems.findIndex(
            (p) => p.problem?.id === problemId,
          );
          if (problemIdx !== -1) {
            setCurrentProblemIndex(problemIdx);
            const sessionProblem = session.problems[problemIdx];
            if (sessionProblem?.problem) {
              setCurrentProblem(sessionProblem.problem);
            }
          }
        } catch (error) {
          console.error("Failed to load session:", error);
          toast.error("Failed to load session");
        }
      } else {
        setIsSessionMode(false);
        if (!currentProblem || currentProblem.id !== problemId) {
          try {
            const problem = await api.problems.getById(problemId);
            setCurrentProblem(problem);
          } catch (error) {
            toast.error("Failed to load problem");
          }
        }
      }
    };

    detectSession();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionIdParam, problemId]);

  useEffect(() => {
    setExecutionResults(null);
  }, [customTestCases]);

  // Poll session data every 5 seconds to update problem generation status
  useEffect(() => {
    if (!isSessionMode || !sessionIdParam) return;

    let isActive = true;

    const pollSession = async () => {
      if (!isActive) return;

      try {
        const session = await api.sessions.get(sessionIdParam as SessionID);

        if (!isActive) return;

        // Update state directly instead of calling loadSession (avoids duplicate API call)
        setSessionData(session);
        const problemIdx = session.problems.findIndex(
          (p) => p.problem?.id === problemId,
        );
        if (problemIdx !== -1) {
          setCurrentProblemIndex(problemIdx);
          const sessionProblem = session.problems[problemIdx];
          if (sessionProblem?.problem) {
            setCurrentProblem(sessionProblem.problem);
          }
        }

        // Stop polling if all problems are ready
        if (session.problems.every((p) => p.status === "ready")) {
          isActive = false;
          return;
        }
      } catch (error) {
        console.error("Session polling error:", error);
      }
    };

    const interval = setInterval(pollSession, 5000);

    return () => {
      isActive = false;
      clearInterval(interval);
    };
  }, [isSessionMode, sessionIdParam, problemId, setCurrentProblem]);

  const handlePreviousProblem = () => {
    if (!sessionData || currentProblemIndex === 0) return;

    const prevProblem = sessionData.problems[currentProblemIndex - 1];
    if (prevProblem?.problem) {
      router.push(
        `/problem/${prevProblem.problem.id}?session=${sessionData.id}`,
      );
    }
  };

  const handleNextProblem = () => {
    if (!sessionData || currentProblemIndex >= sessionData.problems.length - 1)
      return;

    const nextProblem = sessionData.problems[currentProblemIndex + 1];
    if (nextProblem?.problem && nextProblem.status === "ready") {
      router.push(
        `/problem/${nextProblem.problem.id}?session=${sessionData.id}`,
      );
    }
  };

  const handleCompleteSession = async () => {
    if (!sessionIdParam) return;

    try {
      await api.sessions.complete(sessionIdParam as SessionID);
      cleanupSessionCode();
      toast.success("Session completed!");
      router.push("/start");
    } catch (error) {
      console.error("Failed to complete session:", error);
      toast.error("Failed to complete session");
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
    setExecutionError(null);

    try {
      // Format custom cases for the API
      const formattedCustomCases = customTestCases.map((c) => ({
        input: c.input,
        expected_output: "",
      }));

      const response = await api.execution.execute(
        code || "",
        problemId,
        language,
        formattedCustomCases,
        mode,
      );

      if (response.success || response.results) {
        setExecutionResults(response.results);

        // Different success message for Submit vs Run
        if (mode === "submit") {
          if (response.success) {
            toast.success(
              `ACCEPTED! Passed all ${response.total_cases} test cases.`,
            );
          } else {
            // Wrong Answer: Apply Penalty
            const penaltySeconds =
              Number(
                process.env.NEXT_PUBLIC_WRONG_SUBMISSION_PENALTY_MINUTES || 2,
              ) * 60;
            setTimeLeft((prev) => Math.max(0, prev - penaltySeconds));
            toast.error(
              `Wrong Answer. Passed ${response.total_passed}/${response.total_cases} test cases.`,
            );
            toast.warning(`${penaltySeconds / 60} minutes penalty applied`);
          }
        } else {
          if (response.success) {
            toast.success(
              `Run Passed: ${response.total_passed}/${response.total_cases} cases.`,
            );
          } else {
            toast.error(
              `Run Failed: ${response.total_passed}/${response.total_cases} cases.`,
            );
          }
        }
      } else {
        setExecutionError(
          "Execution failed to return results. Please check your code.",
        );
      }
    } catch (error: any) {
      const msg = error.response?.data?.error || "Execution failed";
      setExecutionError(msg);
    } finally {
      setIsRunning(false);
    }
  };

  // --- Hints State ---
  const solutionPenaltyMinutes =
    Number(process.env.NEXT_PUBLIC_SOLUTION_PENALTY_MINUTES) || 5;
  const [isHintsUnlocked, setIsHintsUnlocked] = useState(false);

  // Make Easier/Harder handlers (standalone mode only)
  const handleMakeEasier = async () => {
    if (!currentProblem || isSessionMode) return;

    const newRating = adjustRating(currentProblem.rating, -200);
    const focusArea = currentProblem.focus_area || "general";

    toast.info(`Generating easier problem (rating: ${newRating})...`);

    try {
      const newProblem = await api.problems.generate([focusArea], newRating);
      router.push(`/problem/${newProblem.id}`);
      toast.success("New problem generated!");
    } catch (error) {
      console.error("Failed to generate easier problem:", error);
      toast.error("Failed to generate problem");
    }
  };

  const handleMakeHarder = async () => {
    if (!currentProblem || isSessionMode) return;

    const newRating = adjustRating(currentProblem.rating, +200);
    const focusArea = currentProblem.focus_area || "general";

    toast.info(`Generating harder problem (rating: ${newRating})...`);

    try {
      const newProblem = await api.problems.generate([focusArea], newRating);
      router.push(`/problem/${newProblem.id}`);
      toast.success("New problem generated!");
    } catch (error) {
      console.error("Failed to generate harder problem:", error);
      toast.error("Failed to generate problem");
    }
  };

  // Auto-unlock hints when time runs out (standalone mode only)
  useEffect(() => {
    if (!isSessionMode && timeLeft <= 0 && !isHintsUnlocked) {
      setIsHintsUnlocked(true);
      toast.info("Time's up! Hints have been unlocked.");
    }
  }, [timeLeft, isHintsUnlocked, isSessionMode]);

  // Handle Unlock Hint
  const handleUnlockHint = () => {
    const penalty = solutionPenaltyMinutes * 60;

    if (timeLeft <= penalty) {
      setTimeLeft(0);
      setIsHintsUnlocked(true);
      toast.warning("Time exhausted to unlock hints!");
    } else {
      setTimeLeft((prev) => prev - penalty);
      setIsHintsUnlocked(true);
      toast.success(
        `Hints unlocked (-${solutionPenaltyMinutes}:00 penalty applied)`,
      );
    }
  };

  // Split description and hints
  let descriptionDisplay = "";
  let hintsDisplay = "";

  if (currentProblem?.description) {
    const parts = currentProblem.description.split("## Solution Hints");
    descriptionDisplay = parts[0] || "";
    if (parts.length > 1) {
      hintsDisplay = parts[1] || "";
    }
  }

  // --- Left Content: Problem Details ---
  const LeftContent = !currentProblem ? (
    <div className="space-y-4">
      <Skeleton className="h-8 w-3/4" />
      <Skeleton className="h-32 w-full" />
    </div>
  ) : (
    <div className="flex flex-col h-full -m-4">
      <div className="flex-shrink-0 border-b bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80 sticky top-0 z-10 shadow-sm">
        {/* Top Row: Timers + Metadata */}
        <div className="flex items-center justify-between gap-3 px-4 py-2.5 border-b bg-muted/30">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-sm font-medium">
              <ClockIcon className="h-4 w-4 text-muted-foreground" />
              <span className="font-mono font-bold text-base tabular-nums">
                {Math.floor(timeLeft / 60)}:
                {String(timeLeft % 60).padStart(2, "0")}
              </span>
            </div>
            <div className="h-5 w-px bg-border" />
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <ClockIcon className="h-3.5 w-3.5" />
              <span className="font-mono font-semibold tabular-nums">
                {Math.floor(problemTime / 60)}:
                {String(problemTime % 60).padStart(2, "0")}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {isSessionMode && sessionData && (
              <>
                <Badge variant="outline" className="text-xs font-semibold">
                  {currentProblemIndex + 1}/{sessionData.problem_count}
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  {
                    sessionData.problems.filter((p) => p.status === "ready")
                      .length
                  }{" "}
                  ready
                </Badge>
              </>
            )}
            <Badge variant="secondary" className="text-xs font-medium">
              {currentProblem.focus_area || "General"}
            </Badge>
            {currentProblem.rating && (
              <Badge
                variant="secondary"
                className={`text-xs font-mono font-bold ${getRatingColor(currentProblem.rating).text} ${getRatingColor(currentProblem.rating).bg} border ${getRatingColor(currentProblem.rating).border}`}
              >
                {currentProblem.rating}
              </Badge>
            )}
          </div>
        </div>

        {/* Bottom Row: Title + Navigation */}
        <div className="flex items-center justify-between gap-3 px-4 py-2.5">
          <h2 className="font-bold text-base font-comic truncate flex-1 min-w-0">
            {currentProblem.title}
          </h2>

          <div className="flex items-center gap-1.5 flex-shrink-0">
            {isSessionMode && sessionData ? (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handlePreviousProblem}
                  disabled={currentProblemIndex === 0}
                  className="h-8 w-8 p-0 rounded-md hover:bg-accent disabled:opacity-40"
                >
                  <ChevronLeftIcon className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleNextProblem}
                  disabled={
                    currentProblemIndex >= sessionData.problems.length - 1 ||
                    sessionData.problems[currentProblemIndex + 1]?.status !==
                      "ready"
                  }
                  className="h-8 w-8 p-0 rounded-md hover:bg-accent disabled:opacity-40"
                >
                  <ChevronRightIcon className="h-4 w-4" />
                </Button>
                {currentProblemIndex === sessionData.problems.length - 1 && (
                  <Button
                    size="sm"
                    onClick={handleCompleteSession}
                    disabled={
                      sessionData.problems.filter((p) => p.status === "ready")
                        .length < sessionData.problem_count
                    }
                    className="h-8 px-3 text-xs font-semibold ml-1"
                  >
                    Complete
                  </Button>
                )}
              </>
            ) : (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleMakeEasier}
                  className="h-8 px-3 text-xs font-medium hover:bg-accent"
                >
                  Easier
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleMakeHarder}
                  className="h-8 px-3 text-xs font-medium hover:bg-accent"
                >
                  Harder
                </Button>
                <Button
                  size="sm"
                  onClick={() => router.push("/start")}
                  className="h-8 px-3 text-xs font-semibold ml-1"
                >
                  New
                </Button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Scrollable Content */}
      <div className="flex-1 overflow-auto px-4 py-4">
        {/* Main Description */}
        <div className="prose dark:prose-invert text-sm max-w-none">
          <ReactMarkdown>{descriptionDisplay}</ReactMarkdown>
        </div>

        {/* Hints Section (Standalone Mode Only) */}
        {!isSessionMode && hintsDisplay && (
          <div className="mt-8 pt-6 border-t border-border">
            <h3 className="font-semibold text-sm text-muted-foreground mb-3 flex items-center gap-2 uppercase tracking-wider">
              {isHintsUnlocked ? (
                <UnlockIcon className="w-4 h-4 text-success" />
              ) : (
                <LockIcon className="w-4 h-4 text-warning" />
              )}
              Solution Hints
            </h3>

            {isHintsUnlocked ? (
              <div className="text-sm text-foreground/90 leading-relaxed animate-in fade-in slide-in-from-bottom-2">
                <ReactMarkdown>{hintsDisplay}</ReactMarkdown>
              </div>
            ) : (
              <div className="group relative overflow-hidden rounded-md border bg-muted/30 p-6 text-center transition-all hover:bg-muted/50">
                <div className="absolute inset-0 flex items-center justify-center backdrop-blur-[1px] bg-background/50 pointer-events-none" />

                <div className="relative z-10 flex flex-col items-center gap-3">
                  <div className="p-2 rounded-full bg-background border shadow-sm">
                    <LockIcon className="w-4 h-4 text-muted-foreground" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-foreground">
                      Hints Locked
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Cost:{" "}
                      <span className="font-mono text-destructive">
                        -{solutionPenaltyMinutes}:00
                      </span>
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleUnlockHint}
                    className="h-7 text-xs"
                  >
                    Unlock
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );

  // --- Right Content: Editor & Tests ---
  const RightTop = (
    <CodeEditor
      language={language}
      setLanguage={setLanguage}
      code={code || ""}
      setCode={saveCode}
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
      error={executionError}
    />
  );

  return (
    <div className="h-screen w-screen flex flex-col overflow-hidden bg-background">
      <Navbar
        breadcrumbItems={[
          { label: "home", href: "/" },
          { label: "problem", href: "/problem" },
          { label: currentProblem?.title || "loading..." },
        ]}
      />
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
