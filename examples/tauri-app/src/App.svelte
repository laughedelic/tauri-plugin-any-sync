<script>
  import Greet from './lib/Greet.svelte'
  import Storage from './lib/Storage.svelte'
  import { ping } from 'tauri-plugin-any-sync-api'

	let response = $state('')
  let activeTab = $state('storage') // 'ping' or 'storage'

	function updateResponse(returnValue) {
		response += `[${new Date().toLocaleTimeString()}] ` + (typeof returnValue === 'string' ? returnValue : JSON.stringify(returnValue)) + '<br>'
	}

	function _ping() {
		ping("Pong!").then(updateResponse).catch(updateResponse)
	}
</script>

<main class="container">
  <h1>Tauri AnySync Plugin Demo</h1>

  <div class="tabs">
    <button 
      class:active={activeTab === 'storage'} 
      onclick={() => activeTab = 'storage'}
    >
      Storage Demo
    </button>
    <button 
      class:active={activeTab === 'ping'} 
      onclick={() => activeTab = 'ping'}
    >
      Ping Test
    </button>
  </div>

  {#if activeTab === 'storage'}
    <Storage />
  {:else}
    <div class="ping-section">
      <div class="row">
        <a href="https://vite.dev" target="_blank">
          <img src="/vite.svg" class="logo vite" alt="Vite Logo" />
        </a>
        <a href="https://tauri.app" target="_blank">
          <img src="/tauri.svg" class="logo tauri" alt="Tauri Logo" />
        </a>
        <a href="https://svelte.dev" target="_blank">
          <img src="/svelte.svg" class="logo svelte" alt="Svelte Logo" />
        </a>
      </div>

      <p>
        Click on the Tauri, Vite, and Svelte logos to learn more.
      </p>

      <div class="row">
        <Greet />
      </div>

      <div>
        <button onclick="{_ping}">Ping</button>
        <div>{@html response}</div>
      </div>
    </div>
  {/if}

</main>

<style>
  h1 {
    font-size: 1.75rem;
    font-weight: 600;
    margin: 0 0 1.5rem 0;
    color: #111;
    letter-spacing: -0.02em;
  }

  .tabs {
    display: flex;
    gap: 0.25rem;
    margin: 0 0 2rem 0;
    border-bottom: 1px solid #e5e5e5;
  }

  .tabs button {
    padding: 0.625rem 1.25rem;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-size: 0.9375rem;
    font-weight: 500;
    color: #666;
    transition: all 0.15s;
    letter-spacing: -0.01em;
  }

  .tabs button:hover {
    color: #000;
  }

  .tabs button.active {
    color: #000;
    border-bottom-color: #000;
  }

  .ping-section {
    padding: 1rem 0;
  }

  .logo.vite:hover {
    filter: drop-shadow(0 0 2em #747bff);
  }

  .logo.svelte:hover {
    filter: drop-shadow(0 0 2em #ff3e00);
  }
</style>
