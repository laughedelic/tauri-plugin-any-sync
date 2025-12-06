import { AnimatePresence, motion } from "framer-motion";
import { useNotesStore } from "../../store/useNotesStore";
import { NoteListItem } from "./NoteListItem";
import styles from "./NoteList.module.css";

export function NoteList() {
  const notes = useNotesStore((state) => state.notes);
  const selectedId = useNotesStore((state) => state.selectedId);
  const selectNote = useNotesStore((state) => state.selectNote);

  if (notes.length === 0) {
    return (
      <div className={styles.empty}>
        <p className={styles.emptyText}>No notes yet</p>
        <p className={styles.emptyHint}>Create a note to get started</p>
      </div>
    );
  }

  return (
    <div className={styles.list}>
      <AnimatePresence initial={false} mode="popLayout">
        {notes.map((note) => (
          <motion.div
            key={note.id}
            layout
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.15 }}
          >
            <NoteListItem
              note={note}
              isActive={note.id === selectedId}
              onClick={() => selectNote(note.id)}
            />
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
