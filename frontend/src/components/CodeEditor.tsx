"use client";

import { Button } from "@/components/ui/button";
import { PlayIcon, SendIcon, Loader2Icon } from "lucide-react";

import { useEffect } from "react";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import Editor from "@monaco-editor/react";
import { Skeleton } from "@/components/ui/skeleton";

interface CodeEditorProps {
    language: string;
    setLanguage?: (lang: string) => void;
    code?: string;
    setCode?: (code: string) => void;
    onRun?: () => void;
    onSubmit?: () => void;
    isRunning?: boolean;
    isSubmitting?: boolean;
    readOnly?: boolean;
}

export function CodeEditor({
    language,
    setLanguage,
    code,
    setCode,
    onRun,
    onSubmit,
    isRunning = false,
    isSubmitting = false,
    readOnly = false,
}: CodeEditorProps) {
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.ctrlKey || e.metaKey) {
                if (e.key === "'" || e.code === "Quote") {
                    e.preventDefault();
                    if (!isRunning) onRun?.();
                } else if (e.key === "Enter") {
                    e.preventDefault();
                    if (!isSubmitting) onSubmit?.();
                }
            }
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [onRun, onSubmit, isRunning, isSubmitting]);

    return (
        <div className="h-full flex flex-col">
            <div className="border-b p-2 flex items-center justify-between bg-card flex-shrink-0">
                <div className="flex items-center gap-2">
                    {setLanguage ? (
                        <Select value={language} onValueChange={setLanguage}>
                            <SelectTrigger className="w-[120px] h-8 text-xs bg-background">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="cpp">C++</SelectItem>
                                <SelectItem value="python">Python</SelectItem>
                                <SelectItem value="javascript">JavaScript</SelectItem>
                                {/* Add more as needed */}
                            </SelectContent>
                        </Select>
                    ) : (
                        <span className="text-sm font-medium text-muted-foreground px-2">{language}</span>
                    )}
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="secondary"
                        size="sm"
                        className="h-7 px-3"
                        onClick={onRun}
                        disabled={isRunning || !onRun}
                    >
                        {isRunning ? <Loader2Icon className="h-3 w-3 mr-1.5 animate-spin" /> : <PlayIcon className="h-3 w-3 mr-1.5" />}
                        Run
                    </Button>
                    <Button
                        size="sm"
                        className="h-7 px-3"
                        onClick={onSubmit}
                        disabled={isSubmitting || !onSubmit}
                    >
                        <SendIcon className="h-3 w-3 mr-1.5" />
                        Submit
                    </Button>
                </div>
            </div>
            <div className="flex-1 bg-background relative">
                <Editor
                    height="100%"
                    width="100%"
                    defaultLanguage={language}
                    language={language} // Dynamic update
                    theme="vs-dark"
                    value={code}
                    onChange={(val) => setCode && setCode(val || "")}
                    options={{
                        fontSize: 14,
                        minimap: { enabled: false },
                        scrollBeyondLastLine: false,
                        readOnly: readOnly,
                    }}
                    loading={<Skeleton className="h-full w-full" />}
                />
            </div>
        </div>
    );
}
