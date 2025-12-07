/**
 * Format a date string for display in the notes list.
 * Returns "Today", "Yesterday", or a short date format.
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();

  // Reset time parts for comparison
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const dateOnly = new Date(
    date.getFullYear(),
    date.getMonth(),
    date.getDate()
  );

  if (dateOnly.getTime() === today.getTime()) {
    return "Today";
  }

  if (dateOnly.getTime() === yesterday.getTime()) {
    return "Yesterday";
  }

  // Within the last week, show day name
  const weekAgo = new Date(today);
  weekAgo.setDate(weekAgo.getDate() - 7);
  if (dateOnly.getTime() > weekAgo.getTime()) {
    return date.toLocaleDateString(undefined, { weekday: "long" });
  }

  // Same year, show month and day
  if (date.getFullYear() === now.getFullYear()) {
    return date.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
    });
  }

  // Different year, include year
  return date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

/**
 * Get the first line of content as a preview.
 * Truncates at maxLength characters.
 */
export function getPreview(content: string, maxLength = 80): string {
  if (!content) return "";

  // Get first non-empty line
  const lines = content.split("\n");
  const firstLine = lines.find((line) => line.trim().length > 0) || "";

  if (firstLine.length <= maxLength) {
    return firstLine;
  }

  return firstLine.slice(0, maxLength).trim() + "...";
}
