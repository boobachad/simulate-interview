import { useState, useEffect, useCallback } from "react";
import type { SessionData } from "@/lib/api";

interface SessionTimerState {
  timeLeft: number;
  totalTime: number;
  isRunning: boolean;
}

interface UseSessionTimerReturn {
  timeLeft: number;
  totalTime: number;
  isRunning: boolean;
  startTimer: () => void;
  pauseTimer: () => void;
  resetTimer: () => void;
  setTimeLeft: (time: number | ((prev: number) => number)) => void;
}

const BASE_MINUTES_PER_PROBLEM = 30;

// Rating-based time multipliers (Codeforces-style)
function getRatingMultiplier(rating: number): number {
  if (rating < 1200) return 0.7;  // Easier problems
  if (rating < 1600) return 1.0;  // Medium problems
  if (rating < 2000) return 1.3;  // Hard problems
  return 1.5;                      // Very hard problems
}

function calculateSessionDuration(sessionData: SessionData): number {
  return sessionData.problems.reduce((acc, p) => {
    const rating = p.problem?.rating || 1200;
    const multiplier = getRatingMultiplier(rating);
    return acc + BASE_MINUTES_PER_PROBLEM * multiplier;
  }, 0);
}

function getStorageKey(sessionId: string, key: string): string {
  return `session_${key}_${sessionId}`;
}

export function useSessionTimer(
  sessionId: string | null,
  sessionData: SessionData | null
): UseSessionTimerReturn {
  const [state, setState] = useState<SessionTimerState>(() => {
    if (typeof window === "undefined" || !sessionId) {
      return { timeLeft: 0, totalTime: 0, isRunning: false };
    }

    const storedTimeLeft = sessionStorage.getItem(getStorageKey(sessionId, "timer"));
    const storedTotal = sessionStorage.getItem(getStorageKey(sessionId, "total"));
    const storedStartedAt = sessionStorage.getItem(getStorageKey(sessionId, "started_at"));

    if (storedTimeLeft && storedTotal && storedStartedAt) {
      const startedAt = parseInt(storedStartedAt, 10);
      const elapsed = Math.floor((Date.now() - startedAt) / 1000);
      const totalTime = parseInt(storedTotal, 10);
      const calculatedTimeLeft = Math.max(0, totalTime - elapsed);

      return {
        timeLeft: calculatedTimeLeft,
        totalTime,
        isRunning: calculatedTimeLeft > 0,
      };
    }

    return { timeLeft: 0, totalTime: 0, isRunning: false };
  });

  // Initialize timer when session loads
  useEffect(() => {
    if (!sessionData || !sessionId) return;

    const storedTotal = sessionStorage.getItem(getStorageKey(sessionId, "total"));
    if (storedTotal) return;

    const totalMinutes = calculateSessionDuration(sessionData);
    const totalSeconds = Math.floor(totalMinutes * 60);

    setState({
      timeLeft: totalSeconds,
      totalTime: totalSeconds,
      isRunning: true,
    });

    const now = Date.now();
    sessionStorage.setItem(getStorageKey(sessionId, "total"), totalSeconds.toString());
    sessionStorage.setItem(getStorageKey(sessionId, "timer"), totalSeconds.toString());
    sessionStorage.setItem(getStorageKey(sessionId, "started_at"), now.toString());
  }, [sessionData, sessionId]);

  // Persist timer state to sessionStorage
  useEffect(() => {
    if (!sessionId || state.timeLeft === 0) return;

    sessionStorage.setItem(getStorageKey(sessionId, "timer"), state.timeLeft.toString());
  }, [sessionId, state.timeLeft]);

  // Timer countdown
  useEffect(() => {
    if (!state.isRunning) return;

    const interval = setInterval(() => {
      setState((prev) => {
        if (prev.timeLeft <= 1) {
          return { ...prev, timeLeft: 0, isRunning: false };
        }
        return { ...prev, timeLeft: prev.timeLeft - 1 };
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [state.isRunning]);

  const startTimer = useCallback(() => {
    setState((prev) => ({ ...prev, isRunning: true }));
  }, []);

  const pauseTimer = useCallback(() => {
    setState((prev) => ({ ...prev, isRunning: false }));
  }, []);

  const resetTimer = useCallback(() => {
    if (!sessionId) return;

    setState((prev) => {
      const newTotal = prev.totalTime;
      return {
        timeLeft: newTotal,
        totalTime: newTotal,
        isRunning: true,
      };
    });
    
    const now = Date.now();
    sessionStorage.setItem(getStorageKey(sessionId, "timer"), state.totalTime.toString());
    sessionStorage.setItem(getStorageKey(sessionId, "started_at"), now.toString());
  }, [sessionId, state.totalTime]);

  const setTimeLeft = useCallback(
    (time: number | ((prev: number) => number)) => {
      setState((prev) => {
        const newTime = typeof time === "function" ? time(prev.timeLeft) : time;
        return { ...prev, timeLeft: Math.max(0, newTime) };
      });
    },
    []
  );

  return {
    timeLeft: state.timeLeft,
    totalTime: state.totalTime,
    isRunning: state.isRunning,
    startTimer,
    pauseTimer,
    resetTimer,
    setTimeLeft,
  };
}
