import { useNotesStore } from "../../store/useNotesStore";
import { NoteList } from "../NoteList";
import styles from "./Sidebar.module.css";

export function Sidebar() {
  const notes = useNotesStore((state) => state.notes);
  const createNote = useNotesStore((state) => state.createNote);
  const isLoading = useNotesStore((state) => state.isLoading);

  const handleNewNote = async () => {
    await createNote();
  };

  return (
    <div className={styles.sidebar}>
      <div className={styles.header}>
        <span className={styles.count}>
          {notes.length} {notes.length === 1 ? "note" : "notes"}
        </span>
        <button
          className={styles.newButton}
          onClick={handleNewNote}
          disabled={isLoading}
          aria-label="New note"
        >
          <PlusIcon />
        </button>
      </div>
      <div className={styles.list}>
        <NoteList />
      </div>
    </div>
  );
}

function PlusIcon() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
    >
      <path d="M9 3v12M3 9h12" />
    </svg>
  );
}
