// Rating utility functions for Codeforces-style rating system

export interface RatingColor {
  text: string;
  bg: string;
  border: string;
}

// Get color classes based on rating (Codeforces-style)
export function getRatingColor(rating: number): RatingColor {
  if (rating < 900) {
    return {
      text: "text-gray-600",
      bg: "bg-gray-100",
      border: "border-gray-300",
    };
  } else if (rating < 1200) {
    return {
      text: "text-green-600",
      bg: "bg-green-100",
      border: "border-green-300",
    };
  } else if (rating < 1500) {
    return {
      text: "text-cyan-600",
      bg: "bg-cyan-100",
      border: "border-cyan-300",
    };
  } else if (rating < 1700) {
    return {
      text: "text-blue-600",
      bg: "bg-blue-100",
      border: "border-blue-300",
    };
  } else if (rating < 2000) {
    return {
      text: "text-purple-600",
      bg: "bg-purple-100",
      border: "border-purple-300",
    };
  } else if (rating < 2400) {
    return {
      text: "text-orange-600",
      bg: "bg-orange-100",
      border: "border-orange-300",
    };
  } else {
    return {
      text: "text-red-600",
      bg: "bg-red-100",
      border: "border-red-300",
    };
  }
}

// Get rating label
export function getRatingLabel(rating: number): string {
  if (rating < 900) return "Newbie";
  if (rating < 1200) return "Pupil";
  if (rating < 1500) return "Specialist";
  if (rating < 1700) return "Expert";
  if (rating < 2000) return "Candidate Master";
  if (rating < 2400) return "Master";
  return "Grandmaster";
}

// Adjust rating for Make Easier/Harder buttons
export function adjustRating(currentRating: number, delta: number): number {
  const safeCurrentRating = Number.isFinite(currentRating) ? currentRating : 1200;
  const safeDelta = Number.isFinite(delta) ? delta : 0;
  
  const newRating = safeCurrentRating + safeDelta;
  return Math.max(800, Math.min(3000, newRating));
}
