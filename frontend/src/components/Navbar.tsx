'use client';

import { useRouter } from "next/navigation";
import { Breadcrumb } from "@/components/Breadcrumb";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Button } from "@/components/ui/button";
import { LogOutIcon, SettingsIcon } from "lucide-react";
import { clearAuthToken } from "@/lib/api";
import { toast } from "sonner";

interface NavbarProps {
    breadcrumbItems: { label: string; href?: string }[];
    showAuth?: boolean;
}

export function Navbar({ breadcrumbItems, showAuth = true }: NavbarProps) {
    const router = useRouter();

    const handleLogout = () => {
        clearAuthToken();
        toast.success("Logged out successfully");
        router.push("/");
    };

    const handleSettings = () => {
        router.push("/settings");
    };

    return (
        <div className="h-14 border-b flex items-center justify-between px-6 bg-background shrink-0">
            <Breadcrumb items={breadcrumbItems} />
            <div className="flex items-center gap-2">
                {showAuth && (
                    <>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={handleSettings}
                        >
                            <SettingsIcon className="h-4 w-4" />
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={handleLogout}
                        >
                            <LogOutIcon className="h-4 w-4" />
                        </Button>
                    </>
                )}
                <ThemeToggle />
            </div>
        </div>
    );
}
