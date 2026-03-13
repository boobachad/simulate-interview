import { useState, useEffect, useCallback } from "react";

interface UseProblemTimerReturn {
  problemTime: number;
  resetProblemTimer: () => void;
}

export function useProblemTimer(
  problemId: string | null,
  sessionId: string | null
): UseProblemTimerReturn {
  const [problemTime, setProblemTime] = useState(0);
  const [startTime, setStartTime] = useState<number | null>(null);

  // Reset timer when problem changes
  useEffect(() => {
    if (!problemId) return;

    const key = sessionId
      ? `problem_timer_${sessionId}_${problemId}`
      : `problem_timer_${problemId}`;

    const now = Date.now();

    try {
      const stored = sessionStorage.getItem(key);
      
      if (stored) {
        const savedStart = parseInt(stored, 10);
        const elapsed = Math.floor((now - savedStart) / 1000);
        setProblemTime(elapsed);
        setStartTime(savedStart);
      } else {
        setProblemTime(0);
        setStartTime(now);
        try {
          sessionStorage.setItem(key, now.toString());
        } catch (storageError) {
          console.warn("Failed to save problem timer to sessionStorage:", storageError);
        }
      }
    } catch (error) {
      console.warn("Failed to read problem timer from sessionStorage:", error);
      setProblemTime(0);
      setStartTime(now);
    }
  }, [problemId, sessionId]);

  // Count up every second
  useEffect(() => {
    if (!startTime) return;

    const interval = setInterval(() => {
      const elapsed = Math.floor((Date.now() - startTime) / 1000);
      setProblemTime(elapsed);
    }, 1000);

    return () => clearInterval(interval);
  }, [startTime]);

  const resetProblemTimer = useCallback(() => {
    if (!problemId) return;

    const key = sessionId
      ? `problem_timer_${sessionId}_${problemId}`
      : `problem_timer_${problemId}`;

    const now = Date.now();
    setStartTime(now);
    setProblemTime(0);
    
    try {
      sessionStorage.setItem(key, now.toString());
    } catch (error) {
      console.warn("Failed to save problem timer to sessionStorage:", error);
    }
  }, [sessionId, problemId]);

  return {
    problemTime,
    resetProblemTimer,
  };
}
