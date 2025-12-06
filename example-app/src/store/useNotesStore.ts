import { create } from "zustand";
import { NotesService, type Note } from "../services/notes";
import { appDataDir } from "@tauri-apps/api/path";

export type SaveStatus = "idle" | "saving" | "saved" | "error";

export interface NoteListItem {
  id: string;
  title: string;
  content: string;
  created: string;
  updated?: string;
}

interface NotesState {
  notes: NoteListItem[];
  selectedId: string | null;
  isLoading: boolean;
  isInitialized: boolean;
  saveStatus: SaveStatus;
  error: string | null;

  // Actions
  initialize: () => Promise<void>;
  selectNote: (id: string | null) => void;
  createNote: () => Promise<string | null>;
  updateNote: (id: string, updates: Partial<Note>) => Promise<void>;
  deleteNote: (id: string) => Promise<void>;
  refresh: () => Promise<void>;
  setSaveStatus: (status: SaveStatus) => void;
}

const notesService = new NotesService();

export const useNotesStore = create<NotesState>((set, get) => ({
  notes: [],
  selectedId: null,
  isLoading: false,
  isInitialized: false,
  saveStatus: "idle",
  error: null,

  initialize: async () => {
    if (get().isInitialized || get().isLoading) return;

    set({ isLoading: true, error: null });

    try {
      const appData = await appDataDir();
      const dataDir = `${appData}/any-sync`;
      await notesService.initialize(dataDir);

      // Load existing notes
      await get().refresh();

      // If no notes exist, create example notes for first-time users
      if (get().notes.length === 0) {
        await notesService.createExampleNotes();
        await get().refresh();
      }

      // Select the first note
      const notes = get().notes;
      if (notes.length > 0) {
        set({ selectedId: notes[0].id });
      }

      set({ isInitialized: true, isLoading: false });
    } catch (e) {
      const error = e instanceof Error ? e.message : "Failed to initialize";
      set({ error, isLoading: false });
    }
  },

  selectNote: (id) => {
    set({ selectedId: id, saveStatus: "idle" });
  },

  createNote: async () => {
    const { isInitialized } = get();
    if (!isInitialized) return null;

    try {
      const now = new Date().toISOString();
      const newNote: Note = {
        title: "",
        content: "",
        created: now,
      };

      const id = await notesService.createNote(newNote);

      // Add to local state immediately (optimistic update)
      const noteItem: NoteListItem = {
        id,
        title: "",
        content: "",
        created: now,
      };

      set((state) => ({
        notes: [noteItem, ...state.notes],
        selectedId: id,
        saveStatus: "saved",
      }));

      return id;
    } catch (e) {
      const error = e instanceof Error ? e.message : "Failed to create note";
      set({ error, saveStatus: "error" });
      return null;
    }
  },

  updateNote: async (id, updates) => {
    const { notes, isInitialized } = get();
    if (!isInitialized) return;

    // Find the current note
    const noteIndex = notes.findIndex((n) => n.id === id);
    if (noteIndex === -1) return;

    const currentNote = notes[noteIndex];
    const updatedNote: Note = {
      title: updates.title ?? currentNote.title,
      content: updates.content ?? currentNote.content,
      created: currentNote.created,
    };

    // Optimistic update
    const updatedNoteItem: NoteListItem = {
      ...currentNote,
      ...updates,
      updated: new Date().toISOString(),
    };

    set((state) => ({
      notes: state.notes.map((n) => (n.id === id ? updatedNoteItem : n)),
      saveStatus: "saving",
    }));

    try {
      await notesService.updateNote(id, updatedNote);
      set({ saveStatus: "saved" });
    } catch (e) {
      // Rollback on error
      set((state) => ({
        notes: state.notes.map((n) => (n.id === id ? currentNote : n)),
        saveStatus: "error",
        error: e instanceof Error ? e.message : "Failed to save",
      }));
    }
  },

  deleteNote: async (id) => {
    const { notes, selectedId, isInitialized } = get();
    if (!isInitialized) return;

    const noteIndex = notes.findIndex((n) => n.id === id);
    if (noteIndex === -1) return;

    // Optimistic delete
    const deletedNote = notes[noteIndex];
    const newNotes = notes.filter((n) => n.id !== id);

    // Select adjacent note
    let newSelectedId: string | null = null;
    if (selectedId === id && newNotes.length > 0) {
      const newIndex = Math.min(noteIndex, newNotes.length - 1);
      newSelectedId = newNotes[newIndex].id;
    } else if (selectedId !== id) {
      newSelectedId = selectedId;
    }

    set({ notes: newNotes, selectedId: newSelectedId });

    try {
      await notesService.deleteNote(id);
    } catch (e) {
      // Rollback on error
      set((state) => ({
        notes: [...state.notes.slice(0, noteIndex), deletedNote, ...state.notes.slice(noteIndex)],
        selectedId: id,
        error: e instanceof Error ? e.message : "Failed to delete",
      }));
    }
  },

  refresh: async () => {
    const { isInitialized } = get();
    // Allow refresh during initialization
    if (!isInitialized && !get().isLoading) return;

    try {
      const notesList = await notesService.listNotes();

      // For each note in the list, fetch full content
      const notesWithContent: NoteListItem[] = await Promise.all(
        notesList.map(async (item) => {
          const note = await notesService.getNote(item.id);
          return {
            id: item.id,
            title: note?.title ?? item.title,
            content: note?.content ?? "",
            created: note?.created ?? item.created,
            updated: note?.updated,
          };
        })
      );

      set({ notes: notesWithContent });
    } catch (e) {
      const error = e instanceof Error ? e.message : "Failed to load notes";
      set({ error });
    }
  },

  setSaveStatus: (status) => {
    set({ saveStatus: status });
  },
}));

// Selector for the currently selected note
export const useSelectedNote = () => {
  return useNotesStore((state) => {
    if (!state.selectedId) return null;
    return state.notes.find((n) => n.id === state.selectedId) ?? null;
  });
};
