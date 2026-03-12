import { useState, useEffect, useCallback, useMemo } from "react";
import { BOILERPLATES } from "@/lib/boilerplates";

interface UseCodePersistenceReturn {
  code: string;
  saveCode: (newCode: string) => void;
  cleanupSessionCode: () => void;
}

const STORAGE_QUOTA_WARNING_THRESHOLD = 0.8;
const DEBOUNCE_DELAY_MS = 500;

function getStorageKey(
  sessionId: string | null,
  problemId: string,
  language: string
): string {
  if (sessionId) {
    return `session_code_${sessionId}_${problemId}_${language}`;
  }
  return `code_${problemId}_${language}`;
}

function checkStorageQuota(): void {
  if (typeof navigator === "undefined" || !navigator.storage?.estimate) {
    return;
  }

  navigator.storage
    .estimate()
    .then(({ usage = 0, quota = 0 }) => {
      if (quota > 0 && usage / quota > STORAGE_QUOTA_WARNING_THRESHOLD) {
        console.warn(
          `localStorage usage is at ${Math.round((usage / quota) * 100)}% of quota`
        );
      }
    })
    .catch((error) => {
      console.warn("Failed to estimate storage quota:", error);
    });
}

function safeLocalStorageGet(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch (error) {
    console.warn("Failed to read from localStorage:", error);
    return null;
  }
}

function safeLocalStorageSet(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
    checkStorageQuota();
  } catch (error) {
    console.error("Failed to save to localStorage:", error);
  }
}

function safeLocalStorageRemove(key: string): void {
  try {
    localStorage.removeItem(key);
  } catch (error) {
    console.warn("Failed to remove from localStorage:", error);
  }
}

export function useCodePersistence(
  problemId: string | null,
  language: string,
  sessionId: string | null = null
): UseCodePersistenceReturn {
  const [code, setCode] = useState<string>(() => {
    if (typeof window === "undefined" || !problemId) {
      return BOILERPLATES[language] || "";
    }

    const key = getStorageKey(sessionId, problemId, language);
    const savedCode = safeLocalStorageGet(key);

    return savedCode || BOILERPLATES[language] || "";
  });

  const [pendingSave, setPendingSave] = useState<string | null>(null);

  // Load code when problem or language changes
  useEffect(() => {
    if (!problemId) return;

    const key = getStorageKey(sessionId, problemId, language);
    const savedCode = safeLocalStorageGet(key);

    if (savedCode) {
      setCode(savedCode);
    } else {
      setCode(BOILERPLATES[language] || "");
    }
  }, [problemId, language, sessionId]);

  // Debounced save with useMemo
  const debouncedSave = useMemo(() => {
    let timeoutId: NodeJS.Timeout | null = null;

    return (codeToSave: string) => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }

      timeoutId = setTimeout(() => {
        if (!problemId) return;

        const key = getStorageKey(sessionId, problemId, language);
        safeLocalStorageSet(key, codeToSave);
        setPendingSave(null);
      }, DEBOUNCE_DELAY_MS);
    };
  }, [problemId, language, sessionId]);

  // Trigger debounced save when pendingSave changes
  useEffect(() => {
    if (pendingSave !== null) {
      debouncedSave(pendingSave);
    }
  }, [pendingSave, debouncedSave]);

  // Listen for storage events (multi-tab sync)
  useEffect(() => {
    if (typeof window === "undefined" || !problemId) return;

    const key = getStorageKey(sessionId, problemId, language);

    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === key && event.newValue !== null) {
        setCode(event.newValue);
      }
    };

    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [problemId, language, sessionId]);

  const saveCode = useCallback((newCode: string) => {
    setCode(newCode);
    setPendingSave(newCode);
  }, []);

  const cleanupSessionCode = useCallback(() => {
    if (!sessionId) return;

    try {
      const keysToRemove: string[] = [];

      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key?.startsWith(`session_code_${sessionId}_`)) {
          keysToRemove.push(key);
        }
      }

      keysToRemove.forEach((key) => safeLocalStorageRemove(key));
    } catch (error) {
      console.warn("Failed to cleanup session code:", error);
    }
  }, [sessionId]);

  return {
    code,
    saveCode,
    cleanupSessionCode,
  };
}
