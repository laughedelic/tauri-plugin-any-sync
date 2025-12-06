/**
 * NotesService - Domain service layer demonstrating SyncSpace API usage
 *
 * This shows the recommended pattern:
 * 1. Application defines its own data model (Note interface)
 * 2. Application handles serialization/deserialization
 * 3. Plugin provides generic document storage with opaque bytes
 */

import { syncspace } from "tauri-plugin-any-sync-api";

// Simple logger with prefix
const log = {
  info: (msg: string, ...args: unknown[]) =>
    console.log(`[NotesService] ${msg}`, ...args),
  error: (msg: string, ...args: unknown[]) =>
    console.error(`[NotesService] ${msg}`, ...args),
};

/**
 * Application-specific data model
 * The plugin doesn't know about this - it only stores bytes
 */
export interface Note {
  title: string;
  content: string;
  created: string;
  updated?: string;
}

/**
 * NotesService wraps the SyncSpace API with domain-specific logic
 */
export class NotesService {
  private spaceId: string | null = null;
  private initialized = false;
  private initializing: Promise<void> | null = null;
  private encoder = new TextEncoder();
  private decoder = new TextDecoder();

  /**
   * Check if service is ready
   */
  isReady(): boolean {
    return this.initialized && this.spaceId !== null;
  }

  /**
   * Initialize the service - creates or gets the notes space
   * Safe to call multiple times - will only initialize once
   */
  async initialize(dataDir: string): Promise<void> {
    // Already initialized
    if (this.initialized) {
      return;
    }

    // Initialization in progress - wait for it
    if (this.initializing) {
      return this.initializing;
    }

    // Start initialization
    this.initializing = this.doInitialize(dataDir);
    try {
      await this.initializing;
      this.initialized = true;
    } finally {
      this.initializing = null;
    }
  }

  private async doInitialize(dataDir: string): Promise<void> {
    log.info("init", { dataDir });
    await syncspace.init({
      dataDir,
      networkId: "local",
      deviceId: "example-app",
      config: {},
    });

    log.info("listSpaces");
    const { spaces } = await syncspace.listSpaces();
    const notesSpace = spaces.find((s) => s.name === "notes");

    if (notesSpace) {
      this.spaceId = notesSpace.spaceId;
      log.info("using existing space", { spaceId: this.spaceId });
    } else {
      log.info("createSpace", { name: "notes" });
      const response = await syncspace.createSpace({
        spaceId: "",
        name: "notes",
        metadata: {
          description: "Personal notes storage",
          created: new Date().toISOString(),
        },
      });
      this.spaceId = response.spaceId;
      log.info("created space", { spaceId: this.spaceId });
    }
  }

  /**
   * Create a new note
   */
  async createNote(note: Note): Promise<string> {
    if (!this.spaceId) {
      throw new Error("NotesService not initialized");
    }

    const json = JSON.stringify(note);
    const data = this.encoder.encode(json);

    log.info("createDocument", { title: note.title || "(untitled)" });
    const response = await syncspace.createDocument({
      spaceId: this.spaceId,
      documentId: "",
      collection: "notes",
      data,
      metadata: {
        title: note.title,
        created: note.created,
      },
    });
    log.info("created", { documentId: response.documentId });

    return response.documentId;
  }

  /**
   * Get a note by ID
   */
  async getNote(documentId: string): Promise<Note | null> {
    if (!this.spaceId) {
      throw new Error("NotesService not initialized");
    }

    try {
      log.info("getDocument", { documentId: documentId.slice(0, 8) });
      const response = await syncspace.getDocument({
        spaceId: this.spaceId,
        documentId,
      });

      if (!response.document) {
        log.info("not found");
        return null;
      }

      const json = this.decoder.decode(response.document.data);
      return JSON.parse(json) as Note;
    } catch (error) {
      log.error("getDocument failed", error);
      return null;
    }
  }

  /**
   * Update an existing note
   */
  async updateNote(documentId: string, note: Note): Promise<void> {
    if (!this.spaceId) {
      throw new Error("NotesService not initialized");
    }

    const updatedNote = {
      ...note,
      updated: new Date().toISOString(),
    };

    const json = JSON.stringify(updatedNote);
    const data = this.encoder.encode(json);

    log.info("updateDocument", {
      documentId: documentId.slice(0, 8),
      title: note.title || "(untitled)",
    });
    await syncspace.updateDocument({
      spaceId: this.spaceId,
      documentId,
      data,
      metadata: {
        title: updatedNote.title,
        created: updatedNote.created,
        updated: updatedNote.updated,
      },
      expectedVersion: 0n,
    });
    log.info("updated");
  }

  /**
   * Delete a note
   */
  async deleteNote(documentId: string): Promise<boolean> {
    if (!this.spaceId) {
      throw new Error("NotesService not initialized");
    }

    try {
      log.info("deleteDocument", { documentId: documentId.slice(0, 8) });
      await syncspace.deleteDocument({
        spaceId: this.spaceId,
        documentId,
      });
      log.info("deleted");
      return true;
    } catch (error) {
      log.error("deleteDocument failed", error);
      return false;
    }
  }

  /**
   * List all notes
   */
  async listNotes(): Promise<
    Array<{ id: string; title: string; created: string }>
  > {
    if (!this.spaceId) {
      throw new Error("NotesService not initialized");
    }

    log.info("listDocuments");
    const response = await syncspace.listDocuments({
      spaceId: this.spaceId,
      collection: "notes",
      limit: 0,
      cursor: "",
    });
    log.info("listed", { count: response.documents.length });

    return response.documents.map((doc) => ({
      id: doc.documentId,
      title: doc.metadata["title"] || "Untitled",
      created: doc.metadata["created"] || "",
    }));
  }

  /**
   * Shutdown the service
   */
  async shutdown(): Promise<void> {
    log.info("shutdown");
    await syncspace.shutdown();
    this.spaceId = null;
    this.initialized = false;
  }

  /**
   * Create example notes for first-time users
   */
  async createExampleNotes(): Promise<void> {
    log.info("creating example notes");
    const exampleNotes: Note[] = [
      {
        title: "Welcome to AnySync Notes!",
        content:
          "This is your new notes app. It syncs automatically and works offline.\n\nFeel free to edit or delete these example notes.",
        created: new Date().toISOString(),
      },
      {
        title: "Grocery List",
        content:
          "- Mass-produced cheese\n- Artisanal water\n- Organic air\n- Free-range electrons\n- Gluten-free gluten",
        created: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 min ago
      },
      {
        title: "Meeting Notes",
        content:
          "Discussed synergy. Agreed to circle back. Will leverage our learnings going forward.\n\nAction item: Schedule meeting to discuss next meeting.",
        created: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2 hours ago
      },
      {
        title: "Life Goals",
        content:
          "1. Learn to juggle\n2. Finally read that book\n3. Remember what the book was\n4. Find out where I put my keys\n5. World domination (optional)",
        created: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // Yesterday
      },
      {
        title: "Password Hints",
        content:
          "Email: The name of my first pet + birthday (but clever)\nBank: Something I'll definitely remember\nNetflix: Ask my sister",
        created: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(), // 3 days ago
      },
    ];

    for (const note of exampleNotes) {
      await this.createNote(note);
    }
  }
}
