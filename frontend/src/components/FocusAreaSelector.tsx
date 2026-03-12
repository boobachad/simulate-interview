"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { SearchIcon, Loader2Icon } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { toast } from "sonner";
import type { FocusArea, FocusSelection } from "@/types/api";
import { isSingleTopicMode } from "@/types/api";

interface FocusAreaSelectorProps {
  selection: FocusSelection;
  onSelectionChange: (selection: FocusSelection) => void;
}

export function FocusAreaSelector({
  selection,
  onSelectionChange,
}: FocusAreaSelectorProps) {
  const [focusAreas, setFocusAreas] = useState<FocusArea[]>([]);
  const [filteredAreas, setFilteredAreas] = useState<FocusArea[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const PAGE_SIZE = 20;

  useEffect(() => {
    loadFocusAreas();
  }, []);

  useEffect(() => {
    filterAreas();
  }, [searchQuery, focusAreas]);

  const loadFocusAreas = async () => {
    try {
      const areas = await apiClient.focusAreas.list();
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
      setPage(1);
      return;
    }

    const query = searchQuery.toLowerCase();
    const filtered = focusAreas.filter((area) =>
      area.topic.toLowerCase().includes(query)
    );
    setFilteredAreas(filtered);
    setPage(1);
  };

  const paginatedAreas = filteredAreas.slice(0, page * PAGE_SIZE);
  const hasMore = paginatedAreas.length < filteredAreas.length;

  const handleModeChange = (value: string) => {
    if (value === "all") {
      onSelectionChange({ mode: "all" });
    }
  };

  const handleTopicSelect = (topic: string) => {
    onSelectionChange({ mode: "single", topic });
  };

  return (
    <div className="space-y-4">
      <div>
        <h3 className="font-semibold text-sm mb-2">Focus Mode</h3>
        <RadioGroup
          value={selection.mode}
          onValueChange={handleModeChange}
          className="flex gap-4"
        >
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="all" id="mode-all" />
            <Label htmlFor="mode-all" className="font-normal cursor-pointer">
              All Topics (Random)
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="single" id="mode-single" />
            <Label htmlFor="mode-single" className="font-normal cursor-pointer">
              Single Topic
            </Label>
          </div>
        </RadioGroup>
      </div>

      {isSingleTopicMode(selection) && (
        <div className="space-y-3">
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
              <div className="max-h-[300px] overflow-y-auto space-y-1 border rounded-md p-2">
                {paginatedAreas.map((area) => (
                  <button
                    key={`${area.platform}-${area.topic}`}
                    onClick={() => handleTopicSelect(area.topic)}
                    className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                      selection.topic === area.topic
                        ? "bg-primary text-primary-foreground"
                        : "hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-medium">{area.topic}</span>
                      <div className="flex items-center gap-2 shrink-0">
                        {area.user_solved !== undefined && (
                          <Badge variant="secondary" className="text-xs">
                            {area.user_solved} solved
                          </Badge>
                        )}
                        <Badge variant="outline" className="text-xs">
                          {area.problem_count} problems
                        </Badge>
                      </div>
                    </div>
                  </button>
                ))}
              </div>

              {hasMore && (
                <button
                  onClick={() => setPage((p) => p + 1)}
                  className="w-full text-sm text-muted-foreground hover:text-foreground py-2"
                >
                  Load more...
                </button>
              )}

              {filteredAreas.length === 0 && (
                <p className="text-center text-sm text-muted-foreground py-4">
                  No topics found
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
