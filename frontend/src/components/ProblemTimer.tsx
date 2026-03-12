"use client";

import { TimerIcon } from "lucide-react";

interface ProblemTimerProps {
  problemTime: number;
}

export function ProblemTimer({ problemTime }: ProblemTimerProps) {
  const minutes = Math.floor(problemTime / 60);
  const seconds = problemTime % 60;
  const formattedTime = `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;

  return (
    <div className="flex items-center gap-2 px-3 py-2 rounded-lg border bg-muted/30 text-muted-foreground">
      <TimerIcon className="h-4 w-4" />
      <span className="text-xs font-medium">Problem Time:</span>
      <span className="font-mono text-sm font-semibold">{formattedTime}</span>
    </div>
  );
}
