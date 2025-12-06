import { useState, useEffect, useRef, useCallback } from "react";
import { useNotesStore, useSelectedNote } from "../../store/useNotesStore";
import { useAutoSave } from "../../hooks/useAutoSave";
import { EmptyState } from "../EmptyState";
import styles from "./Editor.module.css";

export function Editor() {
  const selectedNote = useSelectedNote();
  const selectedId = useNotesStore((state) => state.selectedId);
  const deleteNote = useNotesStore((state) => state.deleteNote);
  const saveStatus = useNotesStore((state) => state.saveStatus);

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const titleRef = useRef<HTMLInputElement>(null);

  // Sync local state with selected note
  useEffect(() => {
    if (selectedNote) {
      setTitle(selectedNote.title);
      setContent(selectedNote.content);
    } else {
      setTitle("");
      setContent("");
    }
  }, [selectedNote?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  // Focus title when new empty note is created
  useEffect(() => {
    if (selectedNote && !selectedNote.title && !selectedNote.content) {
      titleRef.current?.focus();
    }
  }, [selectedNote?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-save hook
  useAutoSave({
    noteId: selectedId,
    title,
    content,
    delay: 500,
  });

  // Auto-resize textarea
  const adjustTextareaHeight = useCallback(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    }
  }, []);

  useEffect(() => {
    adjustTextareaHeight();
  }, [content, adjustTextareaHeight]);

  const handleDelete = async () => {
    if (selectedId) {
      await deleteNote(selectedId);
    }
  };

  if (!selectedNote) {
    return <EmptyState type="no-selection" />;
  }

  return (
    <div className={styles.editor}>
      <header className={styles.header}>
        <div className={styles.status}>
          <SaveStatusIndicator status={saveStatus} />
        </div>
        <button
          className={styles.deleteButton}
          onClick={handleDelete}
          aria-label="Delete note"
        >
          <TrashIcon />
        </button>
      </header>

      <div className={styles.content}>
        <input
          ref={titleRef}
          type="text"
          className={styles.titleInput}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          autoComplete="off"
        />
        <textarea
          ref={textareaRef}
          className={styles.bodyInput}
          value={content}
          onChange={(e) => {
            setContent(e.target.value);
            adjustTextareaHeight();
          }}
          placeholder="Start writing..."
        />
      </div>
    </div>
  );
}

function SaveStatusIndicator({ status }: { status: string }) {
  if (status === "saving") {
    return <span className={styles.statusText}>Saving...</span>;
  }
  if (status === "saved") {
    return <span className={`${styles.statusText} ${styles.statusSaved}`}>Saved</span>;
  }
  if (status === "error") {
    return <span className={`${styles.statusText} ${styles.statusError}`}>Error saving</span>;
  }
  return null;
}

function TrashIcon() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M3 4.5h12M6.75 4.5V3a1.5 1.5 0 011.5-1.5h1.5a1.5 1.5 0 011.5 1.5v1.5M14.25 4.5v10.5a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5V4.5h10.5z" />
      <path d="M7.5 7.5v6M10.5 7.5v6" />
    </svg>
  );
}
