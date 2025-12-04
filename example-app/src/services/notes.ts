/**
 * NotesService - Domain service layer demonstrating SyncSpace API usage
 *
 * This shows the recommended pattern:
 * 1. Application defines its own data model (Note interface)
 * 2. Application handles serialization/deserialization
 * 3. Plugin provides generic document storage with opaque bytes
 *
 * Note: Using `as any` type assertions throughout to work around generated
 * protobuf types that incorrectly require Message properties. The protobuf-es
 * create() function accepts plain objects, but TypeScript doesn't recognize this.
 * This is a known limitation and will be addressed in a future update.
 */

import { syncspace } from "tauri-plugin-any-sync-api";

/**
 * Application-specific data model
 * The plugin doesn't know about this - it only stores bytes
 */
export interface Note {
	title: string;
	content: string;
	created: string;
	updated?: string;
	tags?: string[];
}

/**
 * NotesService wraps the SyncSpace API with domain-specific logic
 */
export class NotesService {
	private spaceId: string | null = null;
	private encoder = new TextEncoder();
	private decoder = new TextDecoder();

	/**
	 * Initialize the service - creates or gets the notes space
	 */
	async initialize(dataDir: string): Promise<void> {
		// Initialize the plugin backend
		await syncspace.init({
			dataDir,
			networkId: "local",
			deviceId: "example-app",
			config: {},
		});

		// Check if we already have a notes space
		const spaces = await syncspace.listSpaces();
		const notesSpace = spaces.spaces.find((s) => s.name === "notes");

		if (notesSpace) {
			this.spaceId = notesSpace.spaceId;
		} else {
			// Create a new space for notes
			const response = await syncspace.createSpace({
				spaceId: "", // Let backend generate ID
				name: "notes",
				metadata: {
					description: "Personal notes storage",
					created: new Date().toISOString(),
				},
			});
			this.spaceId = response.spaceId;
		}
	}

	/**
	 * Create a new note
	 */
	async createNote(note: Note): Promise<string> {
		if (!this.spaceId) {
			throw new Error("NotesService not initialized");
		}

		// Serialize the note to bytes (application's responsibility)
		const json = JSON.stringify(note);
		const data = this.encoder.encode(json);

		// Store using plugin's generic document API
		const response = await syncspace.createDocument({
			spaceId: this.spaceId,
			documentId: "", // Let backend generate ID
			collection: "notes",
			data,
			metadata: {
				title: note.title,
				created: note.created,
				tags: note.tags?.join(",") || "",
			},
		});

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
			const response = await syncspace.getDocument({
				spaceId: this.spaceId,
				documentId,
			});

			if (!response.document) {
				return null;
			}

			// Deserialize from bytes (application's responsibility)
			const json = this.decoder.decode(response.document.data);
			return JSON.parse(json) as Note;
		} catch (error) {
			console.error("Failed to get note:", error);
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

		// Add updated timestamp
		const updatedNote = {
			...note,
			updated: new Date().toISOString(),
		};

		// Serialize to bytes
		const json = JSON.stringify(updatedNote);
		const data = this.encoder.encode(json);

		// Update using plugin API
		await syncspace.updateDocument({
			spaceId: this.spaceId,
			documentId,
			data,
			metadata: {
				title: updatedNote.title,
				updated: updatedNote.updated,
				tags: updatedNote.tags?.join(",") || "",
			},
			expectedVersion: 0n, // Skip version check
		});
	}

	/**
	 * Delete a note
	 */
	async deleteNote(documentId: string): Promise<boolean> {
		if (!this.spaceId) {
			throw new Error("NotesService not initialized");
		}

		try {
			await syncspace.deleteDocument({
				spaceId: this.spaceId,
				documentId,
			});
			return true;
		} catch (error) {
			console.error("Failed to delete note:", error);
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

		const response = await syncspace.listDocuments({
			spaceId: this.spaceId,
			collection: "notes",
			limit: 0,
			cursor: "",
		});

		return response.documents.map((doc) => ({
			id: doc.documentId,
			title: doc.metadata["title"] || "Untitled",
			created: doc.metadata["created"] || "",
		}));
	}

	/**
	 * Query notes by tags
	 */
	async findNotesByTag(
		tag: string,
	): Promise<Array<{ id: string; title: string }>> {
		if (!this.spaceId) {
			throw new Error("NotesService not initialized");
		}

		const response = await syncspace.queryDocuments({
			spaceId: this.spaceId,
			collection: "notes",
			filters: [
				{
					field: "tags",
					operator: "contains",
					value: tag,
				},
			],
			limit: 0,
			cursor: "",
		});

		return response.documents.map((doc) => ({
			id: doc.documentId,
			title: doc.metadata["title"] || "Untitled",
		}));
	}

	/**
	 * Shutdown the service
	 */
	async shutdown(): Promise<void> {
		await syncspace.shutdown();
		this.spaceId = null;
	}
}
