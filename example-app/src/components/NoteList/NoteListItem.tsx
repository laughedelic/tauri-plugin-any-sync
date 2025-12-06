import type { NoteListItem as NoteListItemType } from "../../store/useNotesStore";
import { formatDate, getPreview } from "../../lib/formatDate";
import styles from "./NoteList.module.css";

interface NoteListItemProps {
  note: NoteListItemType;
  isActive: boolean;
  onClick: () => void;
}

export function NoteListItem({ note, isActive, onClick }: NoteListItemProps) {
  const displayDate = note.updated || note.created;
  const preview = getPreview(note.content);

  return (
    <button
      className={`${styles.item} ${isActive ? styles.itemActive : ""}`}
      onClick={onClick}
      aria-selected={isActive}
    >
      <div className={styles.itemHeader}>
        <span className={styles.itemTitle}>
          {note.title || "New Note"}
        </span>
        <span className={styles.itemDate}>{formatDate(displayDate)}</span>
      </div>
      {preview && <p className={styles.itemPreview}>{preview}</p>}
    </button>
  );
}
