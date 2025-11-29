//! Integration tests for tauri-plugin-any-sync
//!
//! These tests verify that the plugin correctly integrates with the Tauri app
//! and that all commands work end-to-end with the Go backend.
//!
//! The tests use tauri::test to create a headless app instance, which spawns
//! the Go sidecar process and communicates via gRPC without requiring a GUI.
//! Commands are invoked through IPC to test the actual invocation path.

use serde_json::json;
use tauri::test::{get_ipc_response, MockRuntime};
use tauri_app_lib::create_app_builder;

/// Helper function to create a test app instance with webview for IPC testing
fn create_test_app() -> (tauri::App<MockRuntime>, tauri::WebviewWindow<MockRuntime>, String) {
    // Initialize test logging
    let _ = env_logger::builder()
        .is_test(true)
        .filter_level(log::LevelFilter::Debug)
        .try_init();

    // Create the app using the same builder as production but with MockRuntime
    // Use generate_context to load actual capabilities and config
    let app = create_app_builder::<MockRuntime>()
        .build(tauri::generate_context!())
        .expect("failed to build test app");

    // Get the invoke key for IPC requests
    let invoke_key = app.invoke_key().to_string();

    // Create a webview window for IPC testing
    let webview = tauri::WebviewWindowBuilder::new(&app, "main", Default::default())
        .build()
        .expect("failed to create webview window");

    (app, webview, invoke_key)
}

#[tokio::test]
async fn test_ping_command() {
    let (_app, webview, invoke_key) = create_test_app();

    // Test ping with message
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|ping".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "value": "test message"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "Command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(response["value"], "Echo: test message");
}

#[tokio::test]
async fn test_ping_command_empty_message() {
    let (_app, webview, invoke_key) = create_test_app();

    // Test ping with empty message
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|ping".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {}
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "Command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    // Backend prepends "Echo: " to empty messages too
    assert_eq!(response["value"], "Echo: ");
}

#[tokio::test]
async fn test_storage_put_and_get() {
    let (_app, webview, invoke_key) = create_test_app();

    // Test data
    let test_data = json!({
        "name": "John Doe",
        "age": 30,
        "email": "john@example.com",
        "tags": ["developer", "rust", "tauri"]
    });

    // Put document (send as JSON string like the JS API does)
    let test_data_json = test_data.to_string();
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "users",
                    "id": "user123",
                    "documentJson": test_data_json
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Put command failed: {:?}", res);

    // Get document
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "users",
                    "id": "user123"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    
    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    // Parse the returned JSON string
    let retrieved: serde_json::Value = serde_json::from_str(response["documentJson"].as_str().unwrap()).unwrap();
    // The backend adds the ID to the document, so we check only the fields we care about
    assert_eq!(retrieved["name"], test_data["name"]);
    assert_eq!(retrieved["age"], test_data["age"]);
    assert_eq!(retrieved["email"], test_data["email"]);
    assert_eq!(retrieved["tags"], test_data["tags"]);
}

#[tokio::test]
async fn test_storage_get_nonexistent() {
    let (_app, webview, invoke_key) = create_test_app();

    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "users",
                    "id": "nonexistent_id"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    
    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(response["found"], false, "Should return found=false for nonexistent document");
}

#[tokio::test]
async fn test_storage_delete() {
    let (_app, webview, invoke_key) = create_test_app();

    // Create a document first
    let test_data = json!({"title": "To be deleted"});
    let test_data_json = test_data.to_string();
    
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "temp",
                    "id": "delete_me",
                    "documentJson": test_data_json
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Put command failed: {:?}", res);

    // Verify it exists
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "temp",
                    "id": "delete_me"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(response["found"], true, "Document should exist before deletion");

    // Delete document
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_delete".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "temp",
                    "id": "delete_me"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Delete command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(response["existed"], true, "Delete should return existed=true");

    // Verify it's gone
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "temp",
                    "id": "delete_me"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(response["found"], false, "Document should be deleted");
}

#[tokio::test]
async fn test_storage_list() {
    let (_app, webview, invoke_key) = create_test_app();

    // Create multiple documents
    let documents = [
        ("doc1", json!({"title": "First", "order": 1})),
        ("doc2", json!({"title": "Second", "order": 2})),
        ("doc3", json!({"title": "Third", "order": 3})),
    ];

    // Put all documents
    for (id, doc) in &documents {
        let doc_json = doc.to_string();
        let res = get_ipc_response(
            &webview,
            tauri::webview::InvokeRequest {
                cmd: "plugin:any-sync|storage_put".into(),
                callback: tauri::ipc::CallbackFn(0),
                error: tauri::ipc::CallbackFn(1),
                body: json!({
                    "payload": {
                        "collection": "posts",
                        "id": *id,
                        "documentJson": doc_json
                    }
                }).into(),
                headers: Default::default(),
                url: "tauri://localhost".parse().unwrap(),
                invoke_key: invoke_key.clone(),
            },
        );
        assert!(res.is_ok(), "Put command failed for {}: {:?}", id, res);
    }

    // List all documents in collection
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_list".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "posts"
                }
            }).into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    
    assert!(res.is_ok(), "List command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    
    // Should return an object with "ids" array
    let ids = response["ids"].as_array().expect("Expected ids array in response");
    assert_eq!(ids.len(), 3, "Should have 3 documents");
    
    let id_set: std::collections::HashSet<_> = ids.iter()
        .map(|v| v.as_str().expect("IDs should be strings"))
        .collect();
    
    for (id, _) in &documents {
        assert!(id_set.contains(*id), "Missing document ID: {}", id);
    }
}