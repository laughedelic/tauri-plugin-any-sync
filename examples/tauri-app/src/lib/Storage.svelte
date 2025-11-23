<script>
  import { storagePut, storageGet, storageList } from 'tauri-plugin-any-sync-api'

  // State for form inputs
  let collection = $state('notes')
  let documentId = $state('note1')
  let documentJson = $state('{"title": "My First Note", "content": "Hello, AnyStore!", "timestamp": "2025-11-23"}')
  
  // State for get operation
  let getCollection = $state('notes')
  let getId = $state('note1')
  
  // State for list operation
  let listCollection = $state('notes')
  
  // State for results
  let putResult = $state('')
  let getResult = $state('')
  let listResult = $state('')
  let errorMessage = $state('')

  async function handlePut() {
    errorMessage = ''
    putResult = ''
    try {
      const doc = JSON.parse(documentJson)
      await storagePut(collection, documentId, doc)
      putResult = `✓ Stored document in ${collection}/${documentId}`
    } catch (error) {
      errorMessage = `Put error: ${error.message}`
    }
  }

  async function handleGet() {
    errorMessage = ''
    getResult = ''
    try {
      const doc = await storageGet(getCollection, getId)
      if (doc === null) {
        getResult = `Document not found: ${getCollection}/${getId}`
      } else {
        getResult = `Retrieved from ${getCollection}/${getId}:\n${JSON.stringify(doc, null, 2)}`
      }
    } catch (error) {
      errorMessage = `Get error: ${error.message}`
    }
  }

  async function handleList() {
    errorMessage = ''
    listResult = ''
    try {
      const ids = await storageList(listCollection)
      if (ids.length === 0) {
        listResult = `Collection "${listCollection}" is empty`
      } else {
        listResult = `Found ${ids.length} documents in "${listCollection}":\n${ids.join('\n')}`
      }
    } catch (error) {
      errorMessage = `List error: ${error.message}`
    }
  }

  // Example data templates
  function loadExampleNote() {
    collection = 'notes'
    documentId = 'note1'
    documentJson = JSON.stringify({
      title: "My First Note",
      content: "Hello, AnyStore!",
      timestamp: new Date().toISOString()
    }, null, 2)
  }

  function loadExampleTask() {
    collection = 'tasks'
    documentId = 'task1'
    documentJson = JSON.stringify({
      title: "Test AnyStore integration",
      completed: false,
      priority: "high",
      created: new Date().toISOString()
    }, null, 2)
  }

  function loadExampleUser() {
    collection = 'users'
    documentId = 'user123'
    documentJson = JSON.stringify({
      name: "Alice",
      email: "alice@example.com",
      age: 30,
      preferences: {
        theme: "dark",
        notifications: true
      }
    }, null, 2)
  }
</script>

<div class="storage-demo">
  <h2>AnyStore Storage Demo</h2>
  
  {#if errorMessage}
    <div class="error">{errorMessage}</div>
  {/if}

  <!-- Put Document Section -->
  <section class="operation-section">
    <h3>📝 Put Document</h3>
    <div class="form-group">
      <label for="collection">Collection:</label>
      <input id="collection" type="text" bind:value={collection} placeholder="e.g., notes, tasks, users" />
    </div>
    
    <div class="form-group">
      <label for="documentId">Document ID:</label>
      <input id="documentId" type="text" bind:value={documentId} placeholder="e.g., note1, user123" />
    </div>
    
    <div class="form-group">
      <label for="documentJson">Document JSON:</label>
      <textarea id="documentJson" bind:value={documentJson} rows="6" placeholder='Enter JSON document'></textarea>
    </div>
    
    <div class="button-group">
      <button onclick={handlePut}>Store Document</button>
      <button onclick={loadExampleNote} class="secondary">Load Note Example</button>
      <button onclick={loadExampleTask} class="secondary">Load Task Example</button>
      <button onclick={loadExampleUser} class="secondary">Load User Example</button>
    </div>
    
    {#if putResult}
      <div class="result success">{putResult}</div>
    {/if}
  </section>

  <!-- Get Document Section -->
  <section class="operation-section">
    <h3>🔍 Get Document</h3>
    <div class="form-row">
      <div class="form-group">
        <label for="getCollection">Collection:</label>
        <input id="getCollection" type="text" bind:value={getCollection} placeholder="notes" />
      </div>
      
      <div class="form-group">
        <label for="getId">Document ID:</label>
        <input id="getId" type="text" bind:value={getId} placeholder="note1" />
      </div>
      
      <button onclick={handleGet}>Retrieve Document</button>
    </div>
    
    {#if getResult}
      <div class="result"><pre>{getResult}</pre></div>
    {/if}
  </section>

  <!-- List Documents Section -->
  <section class="operation-section">
    <h3>📋 List Documents</h3>
    <div class="form-row">
      <div class="form-group">
        <label for="listCollection">Collection:</label>
        <input id="listCollection" type="text" bind:value={listCollection} placeholder="notes" />
      </div>
      
      <button onclick={handleList}>List IDs</button>
    </div>
    
    {#if listResult}
      <div class="result"><pre>{listResult}</pre></div>
    {/if}
  </section>

  <!-- Usage Instructions -->
  <section class="info-section">
    <h3>💡 Quick Start</h3>
    <ol>
      <li>Click <strong>"Load Note Example"</strong> to populate sample data</li>
      <li>Click <strong>"Store Document"</strong> to save it to AnyStore</li>
      <li>Click <strong>"Retrieve Document"</strong> to fetch it back</li>
      <li>Click <strong>"List IDs"</strong> to see all documents in the collection</li>
      <li>Try storing multiple documents with different IDs!</li>
    </ol>
    
    <p><em>Note: Documents persist across app restarts and are stored locally on your device.</em></p>
  </section>
</div>

<style>
  .storage-demo {
    width: 100%;
    max-width: 900px;
    margin: 0 auto;
  }

  h2 {
    font-size: 1.75rem;
    font-weight: 600;
    margin: 0 0 2rem 0;
    color: #111;
    letter-spacing: -0.02em;
  }

  .operation-section {
    background: #fafafa;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .operation-section h3 {
    margin: 0 0 1.25rem 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #333;
    letter-spacing: -0.01em;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.4rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #555;
  }

  .form-group input,
  .form-group textarea {
    width: 100%;
    padding: 0.625rem 0.75rem;
    border: 1px solid #d4d4d4;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.9375rem;
    background: #ffffff;
    color: #1a1a1a;
    box-sizing: border-box;
    transition: border-color 0.15s;
  }

  .form-group input:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #000;
  }

  .form-group textarea {
    font-family: 'SF Mono', 'Monaco', 'Menlo', 'Courier New', monospace;
    font-size: 0.875rem;
    resize: vertical;
    line-height: 1.5;
  }

  .form-row {
    display: flex;
    gap: 1rem;
    align-items: flex-end;
    flex-wrap: wrap;
  }

  .form-row .form-group {
    flex: 1;
    min-width: 200px;
  }

  .button-group {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-top: 1rem;
  }

  button {
    padding: 0.625rem 1.125rem;
    background: #000;
    color: #fff;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: background 0.15s;
    letter-spacing: -0.01em;
  }

  button:hover {
    background: #333;
  }

  button.secondary {
    background: #f5f5f5;
    color: #555;
    border: 1px solid #d4d4d4;
    font-size: 0.8125rem;
    padding: 0.5rem 0.875rem;
  }

  button.secondary:hover {
    background: #e5e5e5;
    color: #000;
  }

  .result {
    margin-top: 1.25rem;
    padding: 0.875rem 1rem;
    background: #fafafa;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    color: #333;
  }

  .result.success {
    background: #f0fdf4;
    border-color: #86efac;
    color: #166534;
  }

  .result pre {
    margin: 0;
    font-family: 'SF Mono', 'Monaco', 'Menlo', 'Courier New', monospace;
    font-size: 0.8125rem;
    white-space: pre-wrap;
    word-wrap: break-word;
    line-height: 1.5;
    color: inherit;
  }

  .error {
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 4px;
    padding: 0.875rem 1rem;
    margin-bottom: 1rem;
    color: #991b1b;
    font-size: 0.875rem;
  }

  .info-section {
    background: #fafafa;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 1.5rem;
    margin-top: 2rem;
  }

  .info-section h3 {
    margin: 0 0 1rem 0;
    font-size: 1rem;
    font-weight: 600;
    color: #111;
  }

  .info-section ol {
    margin: 0.75rem 0;
    padding-left: 1.25rem;
    color: #555;
    font-size: 0.9375rem;
  }

  .info-section li {
    margin-bottom: 0.5rem;
    line-height: 1.5;
  }

  .info-section li strong {
    color: #111;
    font-weight: 600;
  }

  .info-section p {
    margin: 1rem 0 0 0;
    font-size: 0.875rem;
    color: #666;
  }
</style>
