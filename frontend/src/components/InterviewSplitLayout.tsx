"use client";

import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area"; // Will need to install or just use div with overflow
import { cn } from "@/lib/utils";

interface InterviewSplitLayoutProps {
    leftContent: React.ReactNode;
    rightTopContent: React.ReactNode;
    rightBottomContent: React.ReactNode;
}

export function InterviewSplitLayout({
    leftContent,
    rightTopContent,
    rightBottomContent,
}: InterviewSplitLayoutProps) {
    return (
        <div className="h-full w-full flex flex-col overflow-hidden bg-background">
            <ResizablePanelGroup direction="horizontal" className="flex-1 w-full min-h-0">
                {/* LEFT PANEL: Sidebar (Settings / Problem Info) */}
                <ResizablePanel defaultSize={25} minSize={20} maxSize={40} className="min-h-0 bg-muted/30">
                    <div className="h-full flex flex-col p-4 gap-4 overflow-y-auto">
                        {leftContent}
                    </div>
                </ResizablePanel>

                <ResizableHandle withHandle />

                {/* RIGHT PANEL: Main Content (Editor & Output) */}
                <ResizablePanel defaultSize={75} className="min-h-0 flex flex-col bg-background">
                    <ResizablePanelGroup direction="vertical" className="flex-1">
                        {/* Top: Editor */}
                        <ResizablePanel defaultSize={70} minSize={30} className="min-h-0 flex flex-col">
                            {rightTopContent}
                        </ResizablePanel>

                        <ResizableHandle withHandle />

                        {/* Bottom: Output / Console */}
                        <ResizablePanel defaultSize={30} minSize={10} className="min-h-0 bg-muted/10">
                            {rightBottomContent}
                        </ResizablePanel>
                    </ResizablePanelGroup>
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}
