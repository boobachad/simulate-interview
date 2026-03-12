import { Activity, Database } from "lucide-react";

interface ConnectionStatusProps {
    apiStatus: "online" | "offline";
    dbStatus: "connected" | "disconnected";
    /** Compact mode for minimal space usage */
    compact?: boolean;
}

export function ConnectionStatus({ apiStatus, dbStatus, compact = false }: ConnectionStatusProps) {
    if (compact) {
        // Minimal indicator - just dots with tooltips
        return (
            <div className="flex items-center gap-2">
                <div
                    className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/30 border border-border/50"
                    title={`API: ${apiStatus === "online" ? "Online" : "Offline"}`}
                >
                    <div className="relative">
                        <span className={`block h-2 w-2 rounded-full transition-colors ${apiStatus === "online" ? "bg-success" : "bg-destructive"}`} />
                        {apiStatus === "online" && (
                            <span className="absolute inset-0 h-2 w-2 rounded-full bg-success animate-ping opacity-75" />
                        )}
                    </div>
                    <Activity className="h-3 w-3 text-muted-foreground" />
                </div>

                <div
                    className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted/30 border border-border/50"
                    title={`Database: ${dbStatus === "connected" ? "Connected" : "Disconnected"}`}
                >
                    <div className="relative">
                        <span className={`block h-2 w-2 rounded-full transition-colors ${dbStatus === "connected" ? "bg-success" : "bg-destructive"}`} />
                        {dbStatus === "connected" && (
                            <span className="absolute inset-0 h-2 w-2 rounded-full bg-success animate-ping opacity-75" />
                        )}
                    </div>
                    <Database className="h-3 w-3 text-muted-foreground" />
                </div>
            </div>
        );
    }

    // Full status display (for footer)
    return (
        <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-3 py-1.5 bg-muted/50 rounded-full border border-border">
                <div className="relative">
                    <span className={`block h-2 w-2 rounded-full transition-colors ${apiStatus === "online" ? "bg-success" : "bg-destructive"}`} />
                    {apiStatus === "online" && (
                        <span className="absolute inset-0 h-2 w-2 rounded-full bg-success animate-ping opacity-75" />
                    )}
                </div>
                <span className="text-xs text-muted-foreground font-medium">
                    API {apiStatus === "online" ? "Online" : "Offline"}
                </span>
            </div>

            <div className="flex items-center gap-2 px-3 py-1.5 bg-muted/50 rounded-full border border-border">
                <div className="relative">
                    <span className={`block h-2 w-2 rounded-full transition-colors ${dbStatus === "connected" ? "bg-success" : "bg-destructive"}`} />
                    {dbStatus === "connected" && (
                        <span className="absolute inset-0 h-2 w-2 rounded-full bg-success animate-ping opacity-75" />
                    )}
                </div>
                <span className="text-xs text-muted-foreground font-medium">
                    DB {dbStatus === "connected" ? "Connected" : "Offline"}
                </span>
            </div>
        </div>
    );
}
