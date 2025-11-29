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
fn create_test_app() -> (
    tauri::App<MockRuntime>,
    tauri::WebviewWindow<MockRuntime>,
    String,
) {
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
            })
            .into(),
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
            })
            .into(),
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
            })
            .into(),
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
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    // Parse the returned JSON string
    let retrieved: serde_json::Value =
        serde_json::from_str(response["documentJson"].as_str().unwrap()).unwrap();
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
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["found"], false,
        "Should return found=false for nonexistent document"
    );
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
            })
            .into(),
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
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["found"], true,
        "Document should exist before deletion"
    );

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
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Delete command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["existed"], true,
        "Delete should return existed=true"
    );

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
            })
            .into(),
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
                })
                .into(),
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
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "List command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();

    // Should return an object with "ids" array
    let ids = response["ids"]
        .as_array()
        .expect("Expected ids array in response");
    assert_eq!(ids.len(), 3, "Should have 3 documents");

    let id_set: std::collections::HashSet<_> = ids
        .iter()
        .map(|v| v.as_str().expect("IDs should be strings"))
        .collect();

    for (id, _) in &documents {
        assert!(id_set.contains(*id), "Missing document ID: {}", id);
    }
}

#[tokio::test]
async fn test_storage_update_existing_document() {
    let (_app, webview, invoke_key) = create_test_app();

    // Create initial document
    let initial_data = json!({"title": "Original Title", "version": 1});
    let initial_json = initial_data.to_string();

    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "documents",
                    "id": "update_test",
                    "documentJson": initial_json
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Put command failed: {:?}", res);

    // Update the document (upsert behavior)
    let updated_data = json!({"title": "Updated Title", "version": 2, "new_field": "added"});
    let updated_json = updated_data.to_string();

    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "documents",
                    "id": "update_test",
                    "documentJson": updated_json
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Update command failed: {:?}", res);

    // Verify the document was updated
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "documents",
                    "id": "update_test"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "Get command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    let retrieved: serde_json::Value =
        serde_json::from_str(response["documentJson"].as_str().unwrap()).unwrap();

    // Verify updated values
    assert_eq!(retrieved["title"], "Updated Title");
    assert_eq!(retrieved["version"], 2);
    assert_eq!(retrieved["new_field"], "added");
}

#[tokio::test]
async fn test_storage_list_empty() {
    let (_app, webview, invoke_key) = create_test_app();

    // List a collection that doesn't exist
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_list".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "nonexistent_collection_xyz"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(res.is_ok(), "List command failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();

    // Should return an object with empty "ids" array
    let ids = response["ids"]
        .as_array()
        .expect("Expected ids array in response");
    assert_eq!(ids.len(), 0, "Empty collection should have 0 documents");
}

#[tokio::test]
async fn test_storage_delete_nonexistent() {
    let (_app, webview, invoke_key) = create_test_app();

    // Try to delete a document that doesn't exist
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_delete".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "temp",
                    "id": "never_existed"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );

    assert!(
        res.is_ok(),
        "Delete command should succeed even for nonexistent documents: {:?}",
        res
    );
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["existed"], false,
        "Delete should return existed=false for nonexistent document"
    );
}

#[tokio::test]
async fn test_multiple_collections() {
    let (_app, webview, invoke_key) = create_test_app();

    // Create documents in different collections with the same ID
    let collection1_data = json!({"collection": "first", "value": "data from collection 1"});
    let collection2_data = json!({"collection": "second", "value": "data from collection 2"});

    // Put document in first collection
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection1",
                    "id": "shared_id",
                    "documentJson": collection1_data.to_string()
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Put to collection1 failed: {:?}", res);

    // Put document in second collection with same ID
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_put".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection2",
                    "id": "shared_id",
                    "documentJson": collection2_data.to_string()
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Put to collection2 failed: {:?}", res);

    // Verify collection1 document is correct
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection1",
                    "id": "shared_id"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get from collection1 failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    let doc1: serde_json::Value =
        serde_json::from_str(response["documentJson"].as_str().unwrap()).unwrap();
    assert_eq!(
        doc1["collection"], "first",
        "Document from collection1 should have its own data"
    );
    assert_eq!(doc1["value"], "data from collection 1");

    // Verify collection2 document is correct and isolated from collection1
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection2",
                    "id": "shared_id"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get from collection2 failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    let doc2: serde_json::Value =
        serde_json::from_str(response["documentJson"].as_str().unwrap()).unwrap();
    assert_eq!(
        doc2["collection"], "second",
        "Document from collection2 should have its own data"
    );
    assert_eq!(doc2["value"], "data from collection 2");

    // Delete from collection1 and verify collection2 is unaffected
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_delete".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection1",
                    "id": "shared_id"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Delete from collection1 failed: {:?}", res);

    // Verify collection1 document is gone
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection1",
                    "id": "shared_id"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get from collection1 failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["found"], false,
        "Document should be deleted from collection1"
    );

    // Verify collection2 document still exists
    let res = get_ipc_response(
        &webview,
        tauri::webview::InvokeRequest {
            cmd: "plugin:any-sync|storage_get".into(),
            callback: tauri::ipc::CallbackFn(0),
            error: tauri::ipc::CallbackFn(1),
            body: json!({
                "payload": {
                    "collection": "collection2",
                    "id": "shared_id"
                }
            })
            .into(),
            headers: Default::default(),
            url: "tauri://localhost".parse().unwrap(),
            invoke_key: invoke_key.clone(),
        },
    );
    assert!(res.is_ok(), "Get from collection2 failed: {:?}", res);
    let response = res.unwrap().deserialize::<serde_json::Value>().unwrap();
    assert_eq!(
        response["found"], true,
        "Document should still exist in collection2"
    );
}
