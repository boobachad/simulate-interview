"use client";

import { useHealthMonitor } from "@/hooks/use-health-monitor";
import { ConnectionStatus } from "@/components/ui/connection-status";

export function AppFooter() {
    const { apiStatus, dbStatus } = useHealthMonitor({ enableNotifications: true });

    return (
        <footer className="w-full border-t border-border bg-card px-6 py-3">
            <div className="flex items-center justify-between">
                {/* Brand */}
                <div className="text-sm font-medium text-muted-foreground">
                    Interview Platform
                </div>

                {/* System Status */}
                <ConnectionStatus apiStatus={apiStatus} dbStatus={dbStatus} />
            </div>
        </footer>
    );
}
