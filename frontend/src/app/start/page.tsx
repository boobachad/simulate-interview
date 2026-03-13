'use client';

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ClockIcon, Loader2Icon, PlayIcon, ListIcon, SearchIcon, RefreshCwIcon, ChevronDownIcon } from "lucide-react";
import { InterviewSplitLayout } from "@/components/InterviewSplitLayout";
import { useInterviewStore } from "@/lib/store";
import { api } from "@/lib/api";
import type { FocusSelection, SessionCreateRequest, ActiveSessionSummary } from "@/lib/api";
import { isSingleTopicMode, isMultipleTopicsMode } from "@/lib/api";
import { toast } from "sonner";
import { CodeEditor } from "@/components/CodeEditor";
import { TestCasesView } from "@/components/TestCasesView";
import { Navbar } from "@/components/Navbar";
import { BOILERPLATES } from "@/lib/boilerplates";
import { useFocusSelection } from "@/hooks/useFocusSelection";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

export default function StartPage() {
  const router = useRouter();
  const { focusAreas, setFocusAreas, selectedFocusAreas, toggleFocusArea, clearSelectedFocusAreas, setCurrentProblem } = useInterviewStore();
  const [isGenerating, setIsGenerating] = useState(false);
  const [mode, setMode] = useState<"playground" | "session">("playground");
  const [weakTopics, setWeakTopics] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeSessions, setActiveSessions] = useState<ActiveSessionSummary[]>([]);
  const [isSessionsOpen, setIsSessionsOpen] = useState(false);

  const { selection: focusSelection, setSelection: setFocusSelection, onTopicToggle } = useFocusSelection({ mode: "all" });


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
    checkSessionRecovery();
    loadWeakTopics();
  }, []);

  const loadWeakTopics = async () => {
    try {
      const stats = await api.stats.get();
      const topicCounts = new Map<string, number>();
      
      if (stats.leetcode?.skills) {
        Object.entries(stats.leetcode.skills).forEach(([tag, skill]) => {
          const normalized = tag.toLowerCase();
          topicCounts.set(normalized, (topicCounts.get(normalized) || 0) + skill.problem_count);
        });
      }
      
      if (stats.codeforces?.tags) {
        Object.entries(stats.codeforces.tags).forEach(([tag, count]) => {
          const normalized = tag.toLowerCase();
          topicCounts.set(normalized, (topicCounts.get(normalized) || 0) + count);
        });
      }
      
      const weakest = Array.from(topicCounts.entries())
        .sort((a, b) => a[1] - b[1])
        .slice(0, 10)
        .map(([name]) => name);
      
      setWeakTopics(weakest);
    } catch (error) {
      console.error('Failed to load weak topics:', error);
    }
  };

  const loadActiveSessions = async () => {
    try {
      const response = await api.sessions.list();
      setActiveSessions(response.sessions || []);
    } catch (error) {
      console.error('Failed to load active sessions:', error);
    }
  };

  // Reload active sessions when switching to session tab
  useEffect(() => {
    if (mode === "session") {
      loadActiveSessions();
    }
  }, [mode]);

  const handleResumeSession = (session: ActiveSessionSummary) => {
    if (session.first_ready_problem_id) {
      router.push(`/problem/${session.first_ready_problem_id}?session=${session.id}`);
    } else {
      toast.error("No problems ready in this session");
    }
  };

  const checkSessionRecovery = async () => {
    // Check for interrupted session creation with expiry
    const recoveryData = sessionStorage.getItem('creating_session');
    if (recoveryData) {
      try {
        const parsed = JSON.parse(recoveryData);
        const EXPIRY_MS = 5 * 60 * 1000; // 5 minutes
        
        if (Date.now() - parsed.timestamp > EXPIRY_MS) {
          // Stale recovery data, ignore and remove
          sessionStorage.removeItem('creating_session');
        } else {
          // Valid recovery data
          toast.error("Session creation was interrupted. Please try again.", {
            duration: 5000,
          });
          sessionStorage.removeItem('creating_session');
        }
      } catch (error) {
        // Invalid JSON, remove
        sessionStorage.removeItem('creating_session');
      }
    }

    // Check for active sessions (handles all interruption cases)
    try {
      const response = await api.sessions.list();
      const sessions = response.sessions || [];
      
      // Update active sessions state
      setActiveSessions(sessions);
      
      if (sessions.length > 0) {
        toast.info(`You have ${sessions.length} active session${sessions.length > 1 ? 's' : ''}. Resume from Session tab.`, {
          duration: 7000,
        });
      }
    } catch (error) {
      console.error('Failed to check active sessions:', error);
    }
  };

  useEffect(() => {
    setExecutionResults(null);
  }, [customTestCases]);

  const loadFocusAreas = async () => {
    try {
      const areas = await api.focusAreas.list();
      setFocusAreas(areas);
    } catch (error) {
      console.error('Failed to load focus areas:', error);
    }
  };

  const handleSessionTopicClick = (topic: string) => {
    onTopicToggle(topic);
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
      const problem = await api.problems.generate(selectedFocusAreas);
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
      const profile = await api.profile.get();
      const problemCount = profile.default_problem_count || 5;

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

      // Store session creation state for recovery
      sessionStorage.setItem('creating_session', JSON.stringify({
        timestamp: Date.now(),
        request
      }));

      const response = await api.sessions.create(request);
      
      // Clear recovery state on success
      sessionStorage.removeItem('creating_session');
      
      router.push(`/problem/${response.first_problem.id}?session=${response.session_id}`);
    } catch (error) {
      console.error(error);
      sessionStorage.removeItem('creating_session');
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

      const response = await api.execution.execute(code || "", "playground", language, formattedCustomCases, mode);

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

      {/* Active Sessions (Session Mode Only) */}
      {mode === "session" && activeSessions.length > 0 && (
        <Collapsible open={isSessionsOpen} onOpenChange={setIsSessionsOpen} className="rounded-lg border bg-card text-card-foreground shadow-sm">
          <CollapsibleTrigger asChild>
            <Button variant="ghost" className="w-full justify-between p-4 h-auto hover:bg-accent">
              <div className="flex items-center gap-2">
                <RefreshCwIcon className="h-4 w-4" />
                <span className="font-semibold text-sm">Active Sessions</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="secondary" className="text-xs">{activeSessions.length}</Badge>
                <ChevronDownIcon className={`h-4 w-4 transition-transform ${isSessionsOpen ? 'rotate-180' : ''}`} />
              </div>
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent className="px-4 pb-4">
            <div className="flex flex-col gap-2 pt-2">
              {activeSessions.map((session) => (
                <div
                  key={session.id}
                  className="flex items-center justify-between p-3 rounded-md border bg-background hover:bg-accent cursor-pointer transition-colors"
                  onClick={() => handleResumeSession(session)}
                >
                  <div className="flex flex-col gap-1">
                    <div className="text-xs font-medium">
                      {session.focus_mode === "all" ? "All Topics" : `${session.focus_mode} mode`}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {session.ready_problems}/{session.problem_count} ready
                    </div>
                  </div>
                  <Button size="sm" variant="outline" onClick={(e) => {
                    e.stopPropagation();
                    handleResumeSession(session);
                  }}>
                    Resume
                  </Button>
                </div>
              ))}
            </div>
          </CollapsibleContent>
        </Collapsible>
      )}

      {/* Focus Area Selection */}
      <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 flex flex-col gap-4 flex-1 min-h-0">
        <div className="flex items-center justify-between flex-shrink-0">
          <h3 className="font-semibold text-sm">Focus Areas</h3>
          <Badge variant="secondary" className="text-xs">
            {mode === "session" 
              ? (focusSelection.mode === "all" 
                  ? "All (Random)" 
                  : focusSelection.mode === "single" 
                    ? "1 selected" 
                    : `${focusSelection.topics?.length || 0} selected`)
              : (selectedFocusAreas.length === 0 ? "Random" : `${selectedFocusAreas.length} selected`)}
          </Badge>
        </div>

        <div className="relative flex-shrink-0">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search topics..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <div className="flex-1 min-h-0 overflow-y-auto">
          <div className="flex flex-wrap gap-2 content-start">
            <div
              className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                mode === "session"
                  ? (focusSelection.mode === "all"
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-secondary text-secondary-foreground hover:bg-secondary/80")
                  : (selectedFocusAreas.length === 0 
                      ? "bg-primary text-primary-foreground border-primary" 
                      : "bg-secondary text-secondary-foreground hover:bg-secondary/80")
              }`}
              onClick={() => {
                if (mode === "session") {
                  setFocusSelection({ mode: "all" });
                } else {
                  clearSelectedFocusAreas();
                }
              }}
            >
              All (Random)
            </div>
            {focusAreas
              .filter(area => !searchQuery || area.topic.toLowerCase().includes(searchQuery.toLowerCase()))
              .map(area => {
                const isWeak = weakTopics.includes(area.topic.toLowerCase());
                const isSelected = mode === "session"
                  ? (focusSelection.mode === "single" && focusSelection.topic === area.topic) ||
                    (focusSelection.mode === "multiple" && focusSelection.topics?.includes(area.topic))
                  : selectedFocusAreas.includes(area.topic);
                
                return (
                  <div
                    key={`${area.platform}:${area.topic}`}
                    onClick={() => {
                      if (mode === "session") {
                        handleSessionTopicClick(area.topic);
                      } else {
                        toggleFocusArea(area.topic);
                      }
                    }}
                    className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                      isSelected
                        ? "bg-primary text-primary-foreground border-primary"
                        : isWeak
                        ? "bg-orange-50/10 text-orange-600 border-orange-500/50 hover:bg-orange-50/20"
                        : "bg-background text-foreground border-input hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    {area.topic}
                  </div>
                );
              })}
          </div>
        </div>
      </div>

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
