import { useState } from "react";
import { toast } from "sonner";
import type { FocusSelection } from "@/lib/api";

export function useFocusSelection(initialSelection: FocusSelection = { mode: "all" }) {
  const [selection, setSelection] = useState<FocusSelection>(initialSelection);

  const onTopicToggle = (topic: string | null) => {
    if (topic === null) {
      setSelection({ mode: "all" });
      return;
    }

    if (selection.mode === "all") {
      setSelection({ mode: "single", topic });
      return;
    }

    if (selection.mode === "single") {
      if (selection.topic === topic) {
        setSelection({ mode: "all" });
      } else {
        setSelection({ mode: "multiple", topics: [selection.topic, topic] });
      }
      return;
    }

    if (selection.mode === "multiple") {
      const isSelected = selection.topics.includes(topic);
      
      if (isSelected) {
        const topics = selection.topics.filter((t) => t !== topic);
        if (topics.length === 0) {
          setSelection({ mode: "all" });
        } else if (topics.length === 1) {
          const singleTopic = topics[0];
          if (singleTopic) {
            setSelection({ mode: "single", topic: singleTopic });
          }
        } else {
          setSelection({ mode: "multiple", topics });
        }
      } else {
        if (selection.topics.length >= 10) {
          toast.error("Maximum 10 topics allowed");
          return;
        }
        const topics = [...selection.topics, topic];
        setSelection({ mode: "multiple", topics });
      }
    }
  };

  return {
    selection,
    setSelection,
    onTopicToggle,
  };
}
