"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { SearchIcon, Loader2Icon } from "lucide-react";
import { api } from "@/lib/api";
import { toast } from "sonner";
import type { FocusArea, FocusSelection } from "@/lib/api";

interface FocusAreaSelectorProps {
  selection: FocusSelection;
  onSelectionChange: (selection: FocusSelection) => void;
  weakTopics?: string[];
}

export function FocusAreaSelector({
  selection,
  onSelectionChange,
  weakTopics = [],
}: FocusAreaSelectorProps) {
  const [focusAreas, setFocusAreas] = useState<FocusArea[]>([]);
  const [filteredAreas, setFilteredAreas] = useState<FocusArea[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadFocusAreas();
  }, []);

  useEffect(() => {
    filterAreas();
  }, [searchQuery, focusAreas]);

  const loadFocusAreas = async () => {
    try {
      const areas = await api.focusAreas.list();
      const sorted = areas.sort((a, b) => b.problem_count - a.problem_count);
      setFocusAreas(sorted);
      setFilteredAreas(sorted);
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to load focus areas";
      toast.error(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const filterAreas = () => {
    if (!searchQuery.trim()) {
      setFilteredAreas(focusAreas);
      return;
    }

    const query = searchQuery.toLowerCase();
    const filtered = focusAreas.filter((area) =>
      area.topic.toLowerCase().includes(query)
    );
    setFilteredAreas(filtered);
  };

  const handleChipClick = (topic: string | null) => {
    if (topic === null) {
      onSelectionChange({ mode: "all" });
      return;
    }

    if (selection.mode === "all") {
      onSelectionChange({ mode: "single", topic });
      return;
    }

    if (selection.mode === "single") {
      if (selection.topic === topic) {
        onSelectionChange({ mode: "all" });
      } else {
        onSelectionChange({ mode: "multiple", topics: [selection.topic, topic] });
      }
      return;
    }

    if (selection.mode === "multiple") {
      const isSelected = selection.topics.includes(topic);
      
      if (isSelected) {
        const topics = selection.topics.filter((t) => t !== topic);
        if (topics.length === 0) {
          onSelectionChange({ mode: "all" });
        } else if (topics.length === 1) {
          const singleTopic = topics[0];
          if (singleTopic) {
            onSelectionChange({ mode: "single", topic: singleTopic });
          }
        } else {
          onSelectionChange({ mode: "multiple", topics });
        }
      } else {
        if (selection.topics.length >= 10) {
          toast.error("Maximum 10 topics allowed");
          return;
        }
        const topics = [...selection.topics, topic];
        onSelectionChange({ mode: "multiple", topics });
      }
    }
  };

  const isTopicSelected = (topic: string): boolean => {
    if (selection.mode === "single") {
      return selection.topic === topic;
    }
    if (selection.mode === "multiple") {
      return selection.topics.includes(topic);
    }
    return false;
  };

  const selectedCount =
    selection.mode === "all"
      ? 0
      : selection.mode === "single"
      ? 1
      : selection.topics.length;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold text-sm">Focus Areas</h3>
        <Badge variant="secondary" className="text-xs">
          {selectedCount === 0 ? "All (Random)" : `${selectedCount} selected`}
        </Badge>
      </div>

      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search topics..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-8">
          <Loader2Icon className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <div
              onClick={() => handleChipClick(null)}
              className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                selection.mode === "all"
                  ? "bg-primary text-primary-foreground border-primary"
                  : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
              }`}
            >
              All (Random)
            </div>
          </div>

          <div className="max-h-[300px] overflow-y-auto border rounded-md p-2">
            <div className="flex flex-wrap gap-2">
              {filteredAreas.map((area) => {
                const selected = isTopicSelected(area.topic);
                const isWeak = weakTopics.includes(area.topic.toLowerCase());
                return (
                  <div
                    key={`${area.platform}-${area.topic}`}
                    onClick={() => handleChipClick(area.topic)}
                    className={`cursor-pointer px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                      selected
                        ? "bg-primary text-primary-foreground border-primary"
                        : isWeak
                        ? "bg-orange-50/10 text-orange-600 border-orange-500/50 hover:bg-orange-50/20"
                        : "bg-background text-foreground border-input hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    {area.topic}
                  </div>
                );
              })}
            </div>
          </div>

          {filteredAreas.length === 0 && (
            <p className="text-center text-sm text-muted-foreground py-4">
              No topics found
            </p>
          )}
        </div>
      )}
    </div>
  );
}
