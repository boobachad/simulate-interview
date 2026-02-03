"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { PlusIcon, XIcon } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { TestCase, ExecutionResult } from "@/lib/store";

interface TestCasesViewProps {
    testCases: TestCase[];
    results: ExecutionResult[] | null;
    customTestCases: { id: string; input: string }[];
    setCustomTestCases: (cases: { id: string; input: string }[]) => void;
}

export function TestCasesView({ testCases, results, customTestCases, setCustomTestCases }: TestCasesViewProps) {
    const [activeTab, setActiveTab] = useState("case0");

    const handleAddTestCase = () => {
        const newId = `custom${Date.now()}`;
        setCustomTestCases([...customTestCases, { id: newId, input: "" }]);
        setActiveTab(newId);
    };

    const handleRemoveTestCase = (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        const newCases = customTestCases.filter(c => c.id !== id);
        setCustomTestCases(newCases);
        if (activeTab === id) {
            setActiveTab("case0");
        }
    };

    const updateCustomInput = (id: string, input: string) => {
        setCustomTestCases(customTestCases.map(c => c.id === id ? { ...c, input } : c));
    };

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
                                ? (result.passed ? "text-green-600" : "text-red-600")
                                : "text-muted-foreground";

                            return (
                                <TabsTrigger
                                    key={`std-${index}`}
                                    value={`case${index}`}
                                    className={`rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 pb-2 pt-2 text-sm ${statusColor} data-[state=active]:text-foreground shadow-none whitespace-nowrap`}
                                >
                                    Case {index + 1}
                                    {result && (
                                        <span className={`ml-2 w-2 h-2 rounded-full ${result.passed ? 'bg-green-500' : 'bg-red-500'}`} />
                                    )}
                                </TabsTrigger>
                            );
                        })}

                        {customTestCases.map((customCase, index) => {
                            const resultIndex = testCases.length + index + 1;
                            const result = results ? results.find(r => r.case_number === resultIndex) : null;
                            const statusColor = result
                                ? (result.passed ? "text-green-600" : "text-red-600")
                                : "text-muted-foreground";

                            return (
                                <TabsTrigger
                                    key={customCase.id}
                                    value={customCase.id}
                                    className={`rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-3 pb-2 pt-2 text-sm ${statusColor} data-[state=active]:text-foreground shadow-none group whitespace-nowrap`}
                                >
                                    Custom {index + 1}
                                    <XIcon
                                        className="ml-2 h-3 w-3 opacity-0 group-hover:opacity-100 hover:text-destructive cursor-pointer transition-opacity"
                                        onClick={(e) => handleRemoveTestCase(customCase.id, e)}
                                    />
                                </TabsTrigger>
                            );
                        })}

                        <Button variant="ghost" size="icon" className="h-6 w-6 ml-1 rounded-sm hover:bg-muted" onClick={handleAddTestCase} title="Add Custom Test Case">
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
                                {/* Input Block */}
                                <div className="space-y-1">
                                    <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Input</div>
                                    <div className="rounded-md border bg-muted/30 p-3 font-mono text-sm whitespace-pre-wrap">
                                        {testCase.input}
                                    </div>
                                </div>

                                {/* Expected Output Block */}
                                <div className="space-y-1">
                                    <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Expected Output</div>
                                    <div className="rounded-md border bg-muted/30 p-3 font-mono text-sm whitespace-pre-wrap">
                                        {testCase.expected_output}
                                    </div>
                                </div>

                                {/* Actual Output Block (Only if executed) */}
                                {result && (
                                    <div className="space-y-1">
                                        <div className="flex items-center gap-2">
                                            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Output</div>
                                            <Badge variant={result.passed ? "default" : "destructive"} className="text-[10px] h-5 px-1.5">
                                                {result.passed ? "Accepted" : "Wrong Answer"}
                                            </Badge>
                                        </div>
                                        <div className={`rounded-md border p-4 font-mono text-sm whitespace-pre-wrap ${result.passed ? "bg-green-500/10 border-green-500/20" : "bg-red-500/10 border-red-500/20"
                                            }`}>
                                            {result.actual_output || result.error || "No output"}
                                        </div>
                                    </div>
                                )}
                            </TabsContent>
                        );
                    })}

                    {/* Custom Cases Content - 2 COLUMNS (Input | Output) */}
                    {customTestCases.map((customCase, index) => {
                        const resultIndex = testCases.length + index + 1;
                        const result = results ? results.find(r => r.case_number === resultIndex) : null;

                        return (
                            <TabsContent key={customCase.id} value={customCase.id} className="flex-1 p-4 m-0 h-full overflow-hidden">
                                <div className="grid grid-cols-2 gap-4 h-full">
                                    {/* Left: Input */}
                                    <div className="space-y-2 flex flex-col h-full">
                                        <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Custom Input</div>
                                        <Textarea
                                            className="font-mono text-sm resize-none flex-1 bg-muted/30 p-3"
                                            placeholder="Enter input here..."
                                            value={customCase.input}
                                            onChange={(e) => updateCustomInput(customCase.id, e.target.value)}
                                        />
                                    </div>

                                    {/* Right: Output */}
                                    <div className="space-y-2 flex flex-col h-full">
                                        <div className="flex items-center gap-2">
                                            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Output</div>
                                            {result && (
                                                <Badge variant={result.passed ? "default" : "destructive"} className="text-[10px] h-5 px-1.5">
                                                    {result.passed ? "Success" : "Error"}
                                                </Badge>
                                            )}
                                        </div>
                                        <div className={`rounded-md border flex-1 p-3 font-mono text-sm whitespace-pre-wrap overflow-auto ${result
                                                ? (result.passed ? "bg-green-500/10 border-green-500/20" : "bg-red-500/10 border-red-500/20")
                                                : "bg-muted/30 text-muted-foreground"
                                            }`}>
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
