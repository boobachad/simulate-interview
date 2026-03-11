import { useEffect, useState, useRef } from "react";
import { toast } from "sonner";

const BASE_URL = "http://api.simulate-interview.localhost:1355";

interface HealthStatus {
    apiStatus: "online" | "offline";
    dbStatus: "connected" | "disconnected";
}

interface UseHealthMonitorOptions {
    /** Show toast notifications when status changes */
    enableNotifications?: boolean;
    /** Polling interval in milliseconds */
    pollInterval?: number;
}

export function useHealthMonitor(options: UseHealthMonitorOptions = {}) {
    const { enableNotifications = true, pollInterval = 5000 } = options;

    const [apiStatus, setApiStatus] = useState<"online" | "offline">("offline");
    const [dbStatus, setDbStatus] = useState<"connected" | "disconnected">("disconnected");

    // Track previous status for change detection
    const prevStatusRef = useRef<HealthStatus>({ apiStatus: "offline", dbStatus: "disconnected" });

    useEffect(() => {
        const checkHealth = async () => {
            try {
                const res = await fetch(`${BASE_URL}/health`);
                const data = await res.json();

                const newApiStatus = "online";
                const newDbStatus = data.db === "connected" ? "connected" : "disconnected";

                // Detect changes and show notifications
                if (enableNotifications) {
                    // API status changed
                    if (prevStatusRef.current.apiStatus !== newApiStatus) {
                        if (newApiStatus === "online") {
                            toast.success("API Connection Restored", {
                                description: "Backend server is now online",
                                duration: 3000,
                            });
                        } else {
                            toast.error("API Connection Lost", {
                                description: "Unable to reach backend server",
                                duration: 5000,
                            });
                        }
                    }

                    // DB status changed
                    if (prevStatusRef.current.dbStatus !== newDbStatus) {
                        if (newDbStatus === "connected") {
                            toast.success("Database Connected", {
                                description: "Database connection established",
                                duration: 3000,
                            });
                        } else {
                            toast.warning("Database Disconnected", {
                                description: "Database connection lost",
                                duration: 5000,
                            });
                        }
                    }
                }

                // Update state
                setApiStatus(newApiStatus);
                setDbStatus(newDbStatus);

                // Store current status for next comparison
                prevStatusRef.current = { apiStatus: newApiStatus, dbStatus: newDbStatus };
            } catch (err) {
                const newApiStatus = "offline";
                const newDbStatus = "disconnected";

                // Show notification only on change
                if (enableNotifications && prevStatusRef.current.apiStatus !== newApiStatus) {
                    toast.error("Connection Error", {
                        description: "Unable to reach the server",
                        duration: 5000,
                    });
                }

                setApiStatus(newApiStatus);
                setDbStatus(newDbStatus);
                prevStatusRef.current = { apiStatus: newApiStatus, dbStatus: newDbStatus };
            }
        };

        checkHealth(); // Initial check

        let interval: NodeJS.Timeout | null = null;

        const startPolling = () => {
            if (!interval) {
                interval = setInterval(checkHealth, pollInterval);
            }
        };

        const stopPolling = () => {
            if (interval) {
                clearInterval(interval);
                interval = null;
            }
        };

        // Handle visibility change - only poll when tab is visible
        const handleVisibilityChange = () => {
            if (document.visibilityState === 'visible') {
                checkHealth(); // Immediate check when becoming visible
                startPolling();
            } else {
                stopPolling();
            }
        };

        // Start polling if page is visible
        if (document.visibilityState === 'visible') {
            startPolling();
        }

        // Listen for visibility changes
        document.addEventListener('visibilitychange', handleVisibilityChange);

        return () => {
            stopPolling();
            document.removeEventListener('visibilitychange', handleVisibilityChange);
        };
    }, [enableNotifications, pollInterval]);

    return { apiStatus, dbStatus };
}
