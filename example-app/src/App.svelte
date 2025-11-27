<script>
  import { storagePut, storageGet, storageDelete, storageList, ping } from 'tauri-plugin-any-sync-api'

  let collection = $state('notes')
  let documentId = $state('note1')
  let documentJson = $state('')
  
  let collections = $state(['notes', 'tasks', 'users', 'settings'])
  let documents = $state([])
  let result = $state('')
  let error = $state('')
  let pingStatus = $state('')
  let initialized = $state(false)
  let documentsListContainer = $state(null)

  // Load example data on mount
  $effect(() => {
    if (!initialized) {
      loadExample()
      // If collections exist, select the first one
      if (collections.length > 0) {
        selectCollection(collections[0])
      }
      initialized = true
    }
  })

  // Scroll active document into view when documentId changes
  $effect(() => {
    documentId
    if (documentsListContainer) {
      // Find the active button and scroll it into view
      const activeButton = documentsListContainer.querySelector('.list-item.active')
      if (activeButton) {
        activeButton.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }
    }
  })

  function loadExample() {
    documentJson = JSON.stringify({
      title: "My First Note",
      content: "Hello, AnyStore!",
      created: new Date().toISOString().split('T')[0]
    }, null, 2)
  }

  async function refreshDocuments() {
    try {
      documents = await storageList(collection)
    } catch (e) {
      documents = []
    }
  }

  async function handleStore() {
    error = ''
    result = ''
    try {
      const doc = JSON.parse(documentJson)
      await storagePut(collection, documentId, doc)
      if (!collections.includes(collection)) {
        collections = [...collections, collection]
      }
      result = `✓ Stored ${collection}/${documentId}`
      await refreshDocuments()
    } catch (e) {
      result = `✗ ${e.message}`
    }
  }

  async function handleRetrieve() {
    error = ''
    result = ''
    try {
      const doc = await storageGet(collection, documentId)
      if (doc === null) {
        result = `✗ Not found: ${collection}/${documentId}`
      } else {
        documentJson = JSON.stringify(doc, null, 2)
        result = `✓ Loaded ${collection}/${documentId}`
      }
    } catch (e) {
      result = `✗ ${e.message}`
    }
  }

  async function handleDelete() {
    error = ''
    result = ''
    try {
      // Find current document index before deletion
      const currentIndex = documents.indexOf(documentId)
      
      const existed = await storageDelete(collection, documentId)
      if (existed) {
        result = `✓ Deleted ${collection}/${documentId}`
        
        // Refresh the UI
        await refreshDocuments()
        
        // Select adjacent document
        const newDocs = await storageList(collection)
        if (newDocs.length > 0) {
          // Try to select the document at the same index, or the previous one, or the first one
          const newIndex = Math.min(currentIndex, newDocs.length - 1)
          documentId = newDocs[newIndex]
          await handleRetrieve()
        } else {
          // No documents left, clear the form
          documentJson = ''
          documentId = ''
        }
      } else {
        result = `✗ Document didn't exist: ${collection}/${documentId}`
      }
    } catch (e) {
      result = `✗ ${e.message}`
    }
  }

  async function selectCollection(name) {
    collection = name
    await refreshDocuments()
    // Load first document if available
    const docs = await storageList(name)
    if (docs.length > 0) {
      documentId = docs[0]
      await handleRetrieve()
    } else {
      documentId = ''
      documentJson = ''
      result = ''
    }
  }

  async function selectDocument(id) {
    documentId = id
    await handleRetrieve()
  }

  function createNew() {
    // Generate random ID
    const randomId = `doc-${Math.random().toString(36).substring(2, 9)}`
    documentId = randomId
    
    // Generate varied content based on collection name
    const templates = {
      notes: {
        title: ["Quick Note", "Important", "Reminder", "Ideas", "Meeting Notes"][Math.floor(Math.random() * 5)],
        content: ["Remember to...", "Key points:", "Follow up on...", "Ideas for..."][Math.floor(Math.random() * 4)],
        created: new Date().toISOString().split('T')[0]
      },
      tasks: {
        title: ["New Task", "Todo Item", "Action Item", "Task"][Math.floor(Math.random() * 4)],
        completed: false,
        priority: ["low", "medium", "high"][Math.floor(Math.random() * 3)],
        created: new Date().toISOString()
      },
      users: {
        name: ["Alice", "Bob", "Charlie", "Diana"][Math.floor(Math.random() * 4)],
        email: `user${Math.floor(Math.random() * 1000)}@example.com`,
        active: true
      }
    }
    
    const template = templates[collection] || templates.notes
    documentJson = JSON.stringify(template, null, 2)
    result = ''
  }

  async function testPing() {
    pingStatus = 'Testing...'
    try {
      await ping("test")
      pingStatus = '✓'
      setTimeout(() => pingStatus = '', 2000)
    } catch (e) {
      pingStatus = `✗ ${e.message}`
    }
  }
</script>

<main class="container">
  <header>
    <h1>AnySync Storage Demo</h1>
    <button class="ping-btn" onclick={testPing} title="Test backend connection">
      {pingStatus || '⚡'}
    </button>
  </header>

  <div class="layout">
    <aside class="sidebar">
      <div class="section">
        <h3>Collections</h3>
        {#if collections.length > 0}
          <div class="list">
            {#each collections as name}
              <button 
                class="list-item" 
                class:active={collection === name}
                onclick={() => selectCollection(name)}
              >
                {name}
              </button>
            {/each}
          </div>
        {:else}
          <p class="empty">No collections yet</p>
        {/if}
      </div>

      <div class="section">
        <h3>Documents in "{collection}"</h3>
        {#if documents.length > 0}
          <div class="list" bind:this={documentsListContainer}>
            {#each documents as id}
              <button 
                class="list-item" 
                class:active={documentId === id}
                onclick={() => selectDocument(id)}
              >
                {id}
              </button>
            {/each}
          </div>
        {:else}
          <p class="empty">Empty collection</p>
        {/if}
      </div>
    </aside>

    <div class="main">
      <div class="form">
        <div class="form-section">
          <div class="input-row">
            <label>
              Collection
              <input type="text" bind:value={collection} placeholder="notes" />
            </label>
            <label>
              Document ID
              <input type="text" bind:value={documentId} placeholder="note1" />
            </label>
          </div>
          
          <label>
            Document Data (JSON)
            <textarea bind:value={documentJson} rows="8" placeholder="Enter JSON..."></textarea>
          </label>
        </div>

        <div class="form-footer">
          <div class="actions">
            <button onclick={handleStore} class="primary">💾 Store Document</button>
            <button onclick={handleDelete} class="danger" disabled={!documentId}>🗑️ Delete Document</button>
            <button onclick={createNew} title="Create new document" class="secondary">➕ New</button>
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

  .ping-btn {
    width: 32px;
    height: 32px;
    min-width: 32px;
    min-height: 32px;
    padding: 0;
    background: #f5f5f5;
    color: #555;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  .ping-btn:hover {
    background: #e5e5e5;
    color: #000;
  }

  .layout {
    display: grid;
    grid-template-columns: 1fr 1fr 2fr;
    grid-template-rows: 1fr auto;
    gap: 1rem;
    flex: 1;
    min-height: 0;
    padding: 1rem;
    overflow: hidden;
  }

  .sidebar {
    display: contents;
  }

  .section {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    overflow: hidden;
    background: #fafafa;
  }

  .section:first-of-type {
    grid-column: 1;
    grid-row: 1;
  }

  .section:nth-of-type(2) {
    grid-column: 2;
    grid-row: 1;
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

  .empty {
    font-size: 0.875rem;
    color: #999;
    font-style: italic;
    margin: 0.5rem;
    padding: 0;
  }

  .main {
    grid-column: 3;
    grid-row: 1 / 3;
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

  .input-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    flex-shrink: 0;
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
      grid-template-rows: auto auto 1fr auto auto;
      gap: 0.75rem;
      padding: 0.75rem;
    }

    .section:first-of-type {
      grid-column: 1;
      grid-row: 1;
      max-height: 120px;
    }

    .section:nth-of-type(2) {
      grid-column: 1;
      grid-row: 2;
      max-height: 120px;
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

    .main {
      grid-column: 1;
      grid-row: 3 / 6;
    }

    .form {
      padding: 1rem;
      gap: 0.75rem;
    }

    .input-row {
      grid-template-columns: 1fr;
      gap: 0.5rem;
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
