"use client";

import { useEffect } from "react";
import { ClockIcon, AlertTriangleIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface InterviewTimerProps {
    timeLeft: number; // in seconds
    setTimeLeft: (time: number) => void;
    totalTime?: number;
}

export function InterviewTimer({
    timeLeft,
    setTimeLeft,
    totalTime = 30 * 60,
}: InterviewTimerProps) {

    useEffect(() => {
        if (timeLeft <= 0) return;

        const interval = setInterval(() => {
            setTimeLeft(timeLeft - 1);
        }, 1000);

        return () => clearInterval(interval);
    }, [timeLeft, setTimeLeft]);

    // Format time as MM:SS
    const minutes = Math.floor(timeLeft / 60);
    const seconds = timeLeft % 60;
    const formattedTime = `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;

    // Visual urgency state
    const isUrgent = timeLeft < 5 * 60; // Less than 5 mins
    const isExpired = timeLeft <= 0;

    return (
        <div
            className={cn(
                "flex items-center justify-between p-3 rounded-lg border shadow-sm transition-colors",
                isExpired
                    ? "state-error"
                    : isUrgent
                        ? "state-warning"
                        : "bg-card border-border text-card-foreground"
            )}
        >
            <div className="flex items-center gap-2">
                {isUrgent || isExpired ? (
                    <AlertTriangleIcon className="h-4 w-4" />
                ) : (
                    <ClockIcon className="h-4 w-4 text-muted-foreground" />
                )}
                <span className="text-sm font-medium">Time Remaining</span>
            </div>
            <div className="font-mono text-xl font-bold tracking-wider">
                {formattedTime}
            </div>
        </div>
    );
}
