import { useEffect, useRef, useCallback } from "react";
import { useNotesStore } from "../store/useNotesStore";

interface UseAutoSaveOptions {
  noteId: string | null;
  title: string;
  content: string;
  delay?: number;
}

export function useAutoSave({
  noteId,
  title,
  content,
  delay = 500,
}: UseAutoSaveOptions): void {
  const updateNote = useNotesStore((state) => state.updateNote);
  const setSaveStatus = useNotesStore((state) => state.setSaveStatus);

  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastSavedRef = useRef<{ title: string; content: string } | null>(null);

  // Track if we've made any changes
  const hasChanges = useCallback(() => {
    if (!lastSavedRef.current) return true;
    return (
      lastSavedRef.current.title !== title ||
      lastSavedRef.current.content !== content
    );
  }, [title, content]);

  useEffect(() => {
    if (!noteId) return;

    // Clear any pending save
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Don't save if no changes
    if (!hasChanges()) return;

    // Schedule save after delay
    timeoutRef.current = setTimeout(async () => {
      try {
        await updateNote(noteId, { title, content });
        lastSavedRef.current = { title, content };
      } catch {
        // Error handling is done in the store
      }
    }, delay);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [noteId, title, content, delay, updateNote, hasChanges]);

  // Reset saved reference when note changes
  useEffect(() => {
    if (noteId) {
      lastSavedRef.current = { title, content };
      setSaveStatus("idle");
    }
  }, [noteId]); // eslint-disable-line react-hooks/exhaustive-deps
}
