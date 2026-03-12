"use client";

import { useState, useEffect } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { PlusIcon, XIcon, AlertTriangleIcon } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { TestCase, ExecutionResult } from "@/lib/store";
import { cn } from "@/lib/utils";

interface TestCasesViewProps {
    testCases: TestCase[];
    results: ExecutionResult[] | null;
    customTestCases: { id: string; input: string }[];
    setCustomTestCases: (cases: { id: string; input: string }[]) => void;
    error?: string | null;
}

export function TestCasesView({ testCases, results, customTestCases, setCustomTestCases, error }: TestCasesViewProps) {
    const [activeTab, setActiveTab] = useState("case0");

    // Ensure we switch tab if active one is deleted or invalid
    useEffect(() => {
        // Logic to validate activeTab could go here, but simple fallback on delete is usually enough
    }, [customTestCases]);

    const handleAddTestCase = () => {
        const newId = `custom${Date.now()}`;
        setCustomTestCases([...customTestCases, { id: newId, input: "" }]);
        setActiveTab(newId);
    };

    const handleRemoveTestCase = (id: string) => {
        const newCases = customTestCases.filter(c => c.id !== id);
        setCustomTestCases(newCases);

        if (activeTab === id) {
            setActiveTab("case0");
        }
    };

    const updateCustomInput = (id: string, input: string) => {
        setCustomTestCases(customTestCases.map(c => c.id === id ? { ...c, input } : c));
    };

    if (error) {
        return (
            <div className="h-full flex flex-col bg-background">
                <div className="border-b px-4 h-9 flex items-center bg-muted/20">
                    <div className="flex items-center gap-2 text-destructive font-medium text-sm">
                        <AlertTriangleIcon className="w-4 h-4" />
                        Execution Error
                    </div>
                </div>
                <div className="flex-1 p-4 overflow-auto">
                    <div className="rounded-md border border-destructive/20 bg-destructive/10 p-4 font-mono text-sm whitespace-pre-wrap text-destructive">
                        {error}
                    </div>
                </div>
            </div>
        );
    }

    if ((!testCases || testCases.length === 0) && customTestCases.length === 0) {
        return (
            <div className="h-full flex flex-col items-center justify-center p-4 text-sm text-muted-foreground gap-2">
                <p>No test cases available.</p>
                <Button variant="outline" size="sm" onClick={handleAddTestCase}>
                    <PlusIcon className="h-4 w-4 mr-2" /> Add Custom Case
                </Button>
            </div>
        );
    }

    return (
        <div className="h-full flex flex-col bg-background">
            <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
                <div className="border-b px-4 bg-muted/20 flex-shrink-0 flex items-center justify-between">
                    <TabsList className="h-9 -mb-px bg-transparent p-0 flex gap-2 overflow-x-auto no-scrollbar items-center">
                        {testCases.map((_, index) => {
                            const result = results ? results.find(r => r.case_number === index + 1) : null;
                            const statusColor = result
                                ? (result.passed ? "text-success" : "text-destructive")
                                : "text-muted-foreground";

                            return (
                                <TabsTrigger
                                    key={`std-${index}`}
                                    value={`case${index}`}
                                    className={cn(
                                        "rounded-none border-b-2 border-transparent px-4 pb-2 pt-2 text-sm shadow-none whitespace-nowrap transition-colors",
                                        "data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:text-foreground",
                                        statusColor
                                    )}
                                >
                                    Case {index + 1}
                                    {result && (
                                        <span className={cn("ml-2 w-2 h-2 rounded-full", result.passed ? 'bg-success' : 'bg-destructive')} />
                                    )}
                                </TabsTrigger>
                            );
                        })}

                        {customTestCases.map((customCase, index) => (
                            <TabsTrigger
                                key={customCase.id}
                                value={customCase.id}
                                className={cn(
                                    "rounded-none border-b-2 border-transparent px-3 pb-2 pt-2 text-sm text-muted-foreground shadow-none group whitespace-nowrap transition-colors",
                                    "data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:text-foreground"
                                )}
                            >
                                Custom {index + 1}
                                <div
                                    role="button"
                                    className="ml-2 opacity-0 group-hover:opacity-100 hover:text-destructive cursor-pointer transition-opacity flex items-center justify-center p-0.5 rounded-sm hover:bg-muted/80 pointer-events-auto"
                                    onClick={(e) => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        handleRemoveTestCase(customCase.id);
                                    }}
                                    onPointerDown={(e) => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                    }}
                                    onMouseDown={(e) => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                    }}
                                >
                                    <XIcon className="h-3 w-3" />
                                </div>
                            </TabsTrigger>
                        ))}

                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 ml-1 rounded-sm hover:bg-muted"
                            onClick={handleAddTestCase}
                            title="Add Custom Test Case"
                        >
                            <PlusIcon className="h-3 w-3" />
                        </Button>
                    </TabsList>
                </div>

                <div className="flex-1 overflow-auto min-h-0 bg-background">
                    {/* Standard Cases Content */}
                    {testCases.map((testCase, index) => {
                        const result = results ? results.find(r => r.case_number === index + 1) : null;

                        return (
                            <TabsContent key={`std-content-${index}`} value={`case${index}`} className="flex-1 p-4 space-y-4 m-0 h-full">
                                <div className="space-y-1">
                                    <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Input</div>
                                    <div className="rounded-md border bg-muted/30 p-3 font-mono text-sm whitespace-pre-wrap">
                                        {testCase.input}
                                    </div>
                                </div>

                                <div className="space-y-1">
                                    <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Expected Output</div>
                                    <div className="rounded-md border bg-muted/30 p-3 font-mono text-sm whitespace-pre-wrap">
                                        {testCase.expected_output}
                                    </div>
                                </div>

                                {result && (
                                    <div className="space-y-1">
                                        <div className="flex items-center gap-2">
                                            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Output</div>
                                            <Badge variant={result.passed ? "default" : "destructive"} className="text-[10px] h-5 px-1.5">
                                                {result.passed ? "Accepted" : "Wrong Answer"}
                                            </Badge>
                                        </div>
                                        <div className={cn(
                                            "rounded-md border p-4 font-mono text-sm whitespace-pre-wrap",
                                            result.passed ? "bg-success/10 border-success/20" : "bg-destructive/10 border-destructive/20"
                                        )}>
                                            {result.actual_output || result.error || "No output"}
                                        </div>
                                    </div>
                                )}
                            </TabsContent>
                        );
                    })}

                    {/* Custom Cases Content */}
                    {customTestCases.map((customCase, index) => {
                        const resultIndex = testCases.length + index + 1;
                        const result = results ? results.find(r => r.case_number === resultIndex) : null;

                        return (
                            <TabsContent key={customCase.id} value={customCase.id} className="flex-1 p-4 m-0 h-full overflow-hidden">
                                <div className="grid grid-cols-2 gap-4 h-full">
                                    <div className="space-y-2 flex flex-col h-full">
                                        <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Custom Input</div>
                                        <Textarea
                                            className="font-mono text-sm resize-none flex-1 bg-muted/30 p-3"
                                            placeholder="Enter input here..."
                                            value={customCase.input}
                                            onChange={(e) => updateCustomInput(customCase.id, e.target.value)}
                                        />
                                    </div>

                                    <div className="space-y-2 flex flex-col h-full">
                                        <div className="flex items-center gap-2">
                                            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Output</div>
                                        </div>
                                        <div className={cn(
                                            "rounded-md border flex-1 p-3 font-mono text-sm whitespace-pre-wrap overflow-auto",
                                            result ? "bg-muted/10 border-border" : "bg-muted/30 text-muted-foreground"
                                        )}>
                                            {result ? (
                                                result.actual_output || result.error || "No output"
                                            ) : (
                                                <div className="h-full flex items-center justify-center italic text-xs">
                                                    Run code to see output
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            </TabsContent>
                        );
                    })}
                </div>
            </Tabs>
        </div>
    );
}
