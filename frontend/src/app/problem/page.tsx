'use client';

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PlusIcon, Loader2Icon, FilterIcon } from "lucide-react";
import { problemsApi } from "@/lib/api";
import { Problem } from "@/lib/store";
import { Navbar } from "@/components/Navbar";

export default function ProblemsIndexPage() {
    const router = useRouter();
    const [problems, setProblems] = useState<Problem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [selectedFocusArea, setSelectedFocusArea] = useState<string | null>(null);

    useEffect(() => {
        loadProblems();
    }, [selectedFocusArea]);

    const loadProblems = async () => {
        setIsLoading(true);
        try {
            const data = await problemsApi.getAll(selectedFocusArea || undefined);
            setProblems(data);
        } catch (error) {
            console.error("Failed to load problems:", error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-background font-sans flex flex-col">
            <Navbar breadcrumbItems={[
                { label: 'home', href: '/' },
                { label: 'problem' }
            ]} />
            <div className="flex-1 w-full p-6 md:p-12">
                <div className="max-w-6xl mx-auto space-y-8">
                    {/* Header */}
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <h1 className="text-3xl font-bold tracking-tight">Problem Bank</h1>
                            <p className="text-muted-foreground mt-1">
                                Review and practice generated problems.
                            </p>
                        </div>
                        <Button onClick={() => router.push('/start')} className="shrink-0">
                            <PlusIcon className="w-4 h-4 mr-2" />
                            Generate New
                        </Button>
                    </div>

                    {/* Filters (Basic for now) */}
                    {/* <div className="flex items-center gap-2">
           Placeholder for focus area filter chips if needed 
        </div> */}

                    {/* Content */}
                    {isLoading ? (
                        <div className="flex justify-center py-12">
                            <Loader2Icon className="w-8 h-8 animate-spin text-muted-foreground" />
                        </div>
                    ) : problems.length === 0 ? (
                        <div className="text-center py-12 border-2 border-dashed rounded-lg bg-muted/20">
                            <h3 className="text-lg font-medium">No problems found</h3>
                            <p className="text-muted-foreground mb-4">Start your first simulated interview!</p>
                            <Button variant="outline" onClick={() => router.push('/start')}>
                                Let's Go
                            </Button>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {problems.map((problem) => (
                                <div
                                    key={problem.id}
                                    onClick={() => router.push(`/problem/${problem.id}`)}
                                    className="group relative cursor-pointer rounded-xl border bg-card p-6 shadow-sm transition-all hover:shadow-md hover:border-primary/50"
                                >
                                    <div className="flex justify-between items-start mb-4">
                                        <Badge variant="secondary" className="text-xs">
                                            {problem.focus_area?.name || "General"}
                                        </Badge>
                                        {/* <span className="text-xs text-muted-foreground">
                    {problem.created_at ? formatDistanceToNow(new Date(problem.created_at), { addSuffix: true }) : ''}
                  </span> */}
                                    </div>

                                    <h3 className="font-semibold text-lg leading-tight mb-2 group-hover:text-primary transition-colors line-clamp-2">
                                        {problem.title}
                                    </h3>

                                    <p className="text-sm text-muted-foreground line-clamp-3 mb-4">
                                        {problem.description?.split("##")[0].replace(/[#*`]/g, "") || "No description preview."}
                                    </p>

                                    <div className="absolute bottom-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <Button size="sm" variant="secondary" className="h-8 text-xs">Practice</Button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
