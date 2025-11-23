<script>
  import { storagePut, storageGet, storageList, ping } from 'tauri-plugin-any-sync-api'

  let collection = $state('notes')
  let documentId = $state('note1')
  let documentJson = $state('')
  
  let collections = $state([])
  let documents = $state([])
  let result = $state('')
  let error = $state('')
  let pingStatus = $state('')

  // Load example data on mount
  $effect(() => {
    loadExample()
    initializeCollections()
  })

  function loadExample() {
    documentJson = JSON.stringify({
      title: "My First Note",
      content: "Hello, AnyStore!",
      created: new Date().toISOString().split('T')[0]
    }, null, 2)
  }

  async function initializeCollections() {
    await refreshCollections()
    // If collections exist, select the first one
    if (collections.length > 0) {
      await selectCollection(collections[0])
    }
  }

  async function refreshCollections() {
    // Try common collection names
    const names = ['notes', 'tasks', 'users', 'settings']
    const found = []
    for (const name of names) {
      try {
        const ids = await storageList(name)
        if (ids.length > 0) {
          found.push(name)
        }
      } catch (e) {
        // Skip errors
      }
    }
    collections = found
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
      result = `✓ Stored ${collection}/${documentId}`
      await refreshCollections()
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
        <div class="section-header">
          <h3>Documents in "{collection}"</h3>
          <button class="new-btn" onclick={createNew} title="Create new document">+</button>
        </div>
        {#if documents.length > 0}
          <div class="list">
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
        
        <div class="actions">
          <button onclick={handleStore} class="primary">💾 Store Document</button>
        </div>

        {#if result}
          <div class="result" class:error={result.startsWith('✗')}>{result}</div>
        {/if}
      </div>
    </div>
  </div>
</main>

<style>
  .container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 2rem 1rem;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  h1 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
    color: #111;
  }

  .ping-btn {
    padding: 0.4rem 0.75rem;
    background: #f5f5f5;
    color: #555;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.15s;
  }

  .ping-btn:hover {
    background: #e5e5e5;
    color: #000;
  }

  .layout {
    display: grid;
    grid-template-columns: 200px 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  .sidebar {
    position: sticky;
    top: 2rem;
  }

  .section {
    margin-bottom: 1.5rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .section h3 {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #666;
    margin: 0;
  }

  .new-btn {
    width: 24px;
    height: 24px;
    padding: 0;
    background: #000;
    color: #fff;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1.125rem;
    font-weight: 600;
    line-height: 1;
    transition: background 0.15s;
  }

  .new-btn:hover {
    background: #333;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .list-item {
    padding: 0.5rem 0.75rem;
    background: #f5f5f5;
    color: #333;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    text-align: left;
    transition: all 0.15s;
  }

  .list-item:hover {
    background: #e5e5e5;
    color: #000;
  }

  .list-item.active {
    background: #000;
    color: #fff;
    border-color: #000;
  }

  .empty {
    font-size: 0.875rem;
    color: #999;
    font-style: italic;
    margin: 0.5rem 0;
  }

  .form {
    background: #fafafa;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 1.5rem;
  }

  .input-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin-bottom: 1rem;
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
    resize: vertical;
    margin-bottom: 1rem;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
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

  .result {
    margin-top: 1rem;
    padding: 0.875rem 1rem;
    background: #f0fdf4;
    border: 1px solid #86efac;
    border-radius: 4px;
    color: #166534;
    font-size: 0.875rem;
  }

  .result.error {
    background: #fef2f2;
    border-color: #fecaca;
    color: #991b1b;
  }
</style>
