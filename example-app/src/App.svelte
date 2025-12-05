<script>
  import { NotesService } from './services/notes'
  import { onMount } from 'svelte'
  import { appDataDir } from '@tauri-apps/api/path'

  const notesService = new NotesService()

  let notes = $state([])
  let selectedNoteId = $state(null)
  let noteTitle = $state('')
  let noteContent = $state('')
  let noteTags = $state('')
  let result = $state('')
  let initialized = $state(false)
  let documentsListContainer = $state(null)

  // Initialize service on mount
  onMount(async () => {
    try {
      const appData = await appDataDir()
      const dataDir = `${appData}/any-sync`
      await notesService.initialize(dataDir)
      await refreshNotes()
      initialized = true
      result = '✓ Plugin initialized'
    } catch (e) {
      result = `✗ Initialization failed: ${e.message}`
      console.error(e)
    }
  })

  // Scroll active document into view when selection changes
  $effect(() => {
    selectedNoteId
    if (documentsListContainer) {
      const activeButton = documentsListContainer.querySelector('.list-item.active')
      if (activeButton) {
        activeButton.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }
    }
  })

  async function refreshNotes() {
    try {
      console.log('[App] Refreshing notes...')
      const fetchedNotes = await notesService.listNotes()
      console.log('[App] Fetched notes:', fetchedNotes)
      console.log('[App] Fetched notes length:', fetchedNotes.length)
      console.log('[App] Before assignment - notes:', notes)
      notes = fetchedNotes
      console.log('[App] After assignment - notes:', notes)
      console.log('[App] After assignment - notes.length:', notes.length)
    } catch (e) {
      notes = []
      console.error('Failed to refresh notes:', e)
    }
  }

  async function selectNote(id) {
    try {
      selectedNoteId = id
      const note = await notesService.getNote(id)
      if (note) {
        noteTitle = note.title
        noteContent = note.content
        noteTags = note.tags?.join(', ') || ''
        result = `✓ Loaded note: ${note.title}`
      }
    } catch (e) {
      result = `✗ ${e.message}`
    }
  }

  async function handleSave() {
    console.log('[App] handleSave called')
    result = ''
    try {
      const tags = noteTags.split(',').map(t => t.trim()).filter(t => t)
      console.log('[App] selectedNoteId:', selectedNoteId)
      console.log('[App] noteTitle:', noteTitle)
      console.log('[App] noteContent length:', noteContent.length)

      if (selectedNoteId) {
        // Update existing note
        console.log('[App] Updating existing note')
        await notesService.updateNote(selectedNoteId, {
          title: noteTitle || 'Untitled',
          content: noteContent,
          created: notes.find(n => n.id === selectedNoteId)?.created || new Date().toISOString(),
          tags
        })
        result = `✓ Updated: ${noteTitle}`
      } else {
        // Create new note
        console.log('[App] Creating new note')
        const id = await notesService.createNote({
          title: noteTitle || 'Untitled',
          content: noteContent,
          created: new Date().toISOString(),
          tags
        })
        console.log('[App] Created note with ID:', id)
        selectedNoteId = id
        result = `✓ Created: ${noteTitle}`
      }

      console.log('[App] Calling refreshNotes after save')
      await refreshNotes()
      console.log('[App] refreshNotes completed')
    } catch (e) {
      console.error('[App] Error in handleSave:', e)
      result = `✗ ${e.message}`
    }
  }

  async function handleDelete() {
    if (!selectedNoteId) return
    
    result = ''
    try {
      const currentIndex = notes.findIndex(n => n.id === selectedNoteId)
      const deleted = await notesService.deleteNote(selectedNoteId)
      
      if (deleted) {
        result = `✓ Deleted note`
        await refreshNotes()
        
        // Select adjacent note
        if (notes.length > 0) {
          const newIndex = Math.min(currentIndex, notes.length - 1)
          await selectNote(notes[newIndex].id)
        } else {
          // No notes left, clear form
          selectedNoteId = null
          noteTitle = ''
          noteContent = ''
          noteTags = ''
        }
      }
    } catch (e) {
      result = `✗ ${e.message}`
    }
  }

  function createNew() {
    selectedNoteId = null
    noteTitle = ''
    noteContent = ''
    noteTags = ''
    result = ''
  }
</script>

<main class="container">
  <header>
    <h1>SyncSpace Notes Demo</h1>
    <div class="status">
      {#if initialized}
        <span class="status-badge success">✓ Connected</span>
      {:else}
        <span class="status-badge">⏳ Initializing...</span>
      {/if}
    </div>
  </header>

  <div class="layout">
    <aside class="sidebar">
      <div class="section">
        <h3>All Notes ({notes.length})</h3>
        {#if notes.length > 0}
          <div class="list" bind:this={documentsListContainer}>
            {#each notes as note}
              <button 
                class="list-item" 
                class:active={selectedNoteId === note.id}
                onclick={() => selectNote(note.id)}
              >
                <div class="note-item">
                  <div class="note-title">{note.title}</div>
                  <div class="note-date">{new Date(note.created).toLocaleDateString()}</div>
                </div>
              </button>
            {/each}
          </div>
        {:else}
          <p class="empty">No notes yet. Create one!</p>
        {/if}
      </div>
    </aside>

    <div class="main">
      <div class="form">
        <div class="form-section">
          <label>
            Title
            <input type="text" bind:value={noteTitle} placeholder="Note title" />
          </label>
          
          <label>
            Content
            <textarea bind:value={noteContent} rows="10" placeholder="Start writing..."></textarea>
          </label>

          <label>
            Tags (comma-separated)
            <input type="text" bind:value={noteTags} placeholder="personal, work, ideas" />
          </label>
        </div>

        <div class="form-footer">
          <div class="actions">
            <button onclick={handleSave} class="primary">
              {selectedNoteId ? '💾 Save' : '➕ Create Note'}
            </button>
            <button onclick={handleDelete} class="danger" disabled={!selectedNoteId}>
              🗑️ Delete
            </button>
            <button onclick={createNew} class="secondary">
              📄 New Note
            </button>
          </div>

          {#if result}
            <div class="result" class:error={result.startsWith('✗')}>{result}</div>
          {/if}
        </div>
      </div>
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }

  .container {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    overflow: hidden;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: #fff;
    border-bottom: 1px solid #e5e5e5;
    flex-shrink: 0;
  }

  h1 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
    color: #111;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-badge {
    padding: 0.375rem 0.75rem;
    background: #f5f5f5;
    color: #666;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .status-badge.success {
    background: #f0fdf4;
    color: #166534;
    border-color: #86efac;
  }

  .layout {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 1rem;
    flex: 1;
    min-height: 0;
    padding: 1rem;
    overflow: hidden;
  }

  .sidebar {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .section {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    overflow: hidden;
    background: #fafafa;
    flex: 1;
  }

  .section h3 {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #666;
    margin: 0;
    padding: 0.75rem;
    border-bottom: 1px solid #e5e5e5;
    flex-shrink: 0;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    overflow-y: auto;
    min-height: 0;
    padding: 0.5rem;
  }

  .list-item {
    padding: 0.5rem 0.75rem;
    background: #fff;
    color: #333;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    text-align: left;
    transition: all 0.15s;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    scroll-margin: 0.5rem;
  }

  .list-item:hover {
    background: #e5e5e5;
    color: #000;
  }

  .list-item.active {
    background: #000;
    color: #fff;
    border-color: #000;
    font-weight: 600;
    scroll-margin: 0.5rem;
  }

  .note-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .note-title {
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .note-date {
    font-size: 0.75rem;
    opacity: 0.6;
  }

  .list-item.active .note-date {
    opacity: 0.8;
  }

  .empty {
    font-size: 0.875rem;
    color: #999;
    font-style: italic;
    margin: 0.5rem;
    padding: 0;
  }

  .main {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    min-height: 0;
  }

  .form {
    background: #fafafa;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    flex-shrink: 0;
  }

  .form-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    flex-shrink: 0;
  }

  .form-footer {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e5e5;
  }

  label {
    display: block;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #666;
    margin-bottom: 0.4rem;
  }

  input, textarea {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.875rem;
    background: #fff;
    color: #1a1a1a;
    box-sizing: border-box;
    margin-top: 0.25rem;
  }

  input:focus, textarea:focus {
    outline: none;
    border-color: #000;
  }

  textarea {
    font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
    resize: none;
    flex: 1;
    min-height: 120px;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    flex-shrink: 0;
  }

  button {
    padding: 0.625rem 1.125rem;
    background: #fff;
    color: #333;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.15s;
    white-space: nowrap;
  }

  button:hover {
    background: #f5f5f5;
    border-color: #000;
    color: #000;
  }

  button.primary {
    background: #000;
    color: #fff;
    border-color: #000;
  }

  button.primary:hover {
    background: #333;
  }

  button.secondary {
    background: #6b7280;
    color: #fff;
    border-color: #6b7280;
  }

  button.secondary:hover {
    background: #4b5563;
    border-color: #4b5563;
  }

  button.danger {
    background: #dc2626;
    color: #fff;
    border-color: #dc2626;
  }

  button.danger:hover:not(:disabled) {
    background: #b91c1c;
    border-color: #b91c1c;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .result {
    margin: 0;
    padding: 0.875rem 1rem;
    background: #f0fdf4;
    border: 1px solid #86efac;
    border-radius: 4px;
    color: #166534;
    font-size: 0.875rem;
    flex-shrink: 0;
  }

  .result.error {
    background: #fef2f2;
    border-color: #fecaca;
    color: #991b1b;
  }

  /* Mobile Responsive Layout */
  @media (max-width: 768px) {
    header {
      padding: 0.75rem;
    }

    h1 {
      font-size: 1.1rem;
    }

    .layout {
      grid-template-columns: 1fr;
      grid-template-rows: auto 1fr;
      gap: 0.75rem;
      padding: 0.75rem;
    }

    .sidebar {
      max-height: 200px;
    }

    .section h3 {
      font-size: 0.7rem;
      padding: 0.5rem;
    }

    .list {
      padding: 0.375rem;
      gap: 1px;
    }

    .list-item {
      padding: 0.5rem;
      font-size: 0.75rem;
      min-height: 36px;
      display: flex;
      align-items: center;
    }

    .form {
      padding: 1rem;
      gap: 0.75rem;
    }

    label {
      font-size: 0.65rem;
      margin-bottom: 0.25rem;
    }

    input, textarea {
      padding: 0.5rem;
      font-size: 0.8rem;
      margin-top: 0.15rem;
    }

    textarea {
      min-height: 100px;
      max-height: 150px;
    }

    button {
      padding: 0.5rem 0.875rem;
      font-size: 0.75rem;
      flex: 1;
      min-height: 40px;
    }

    .result {
      padding: 0.625rem 0.75rem;
      font-size: 0.75rem;
    }
  }
</style>
