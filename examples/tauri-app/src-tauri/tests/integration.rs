//! Integration tests for tauri-plugin-any-sync
//!
//! These tests verify that the plugin correctly integrates with the Tauri app
//! and that all commands work end-to-end with the Go backend.
//!
//! The tests use tauri::test to create a headless app instance, which spawns
//! the Go sidecar process and communicates via gRPC without requiring a GUI.

use serde_json::json;
use tauri::test::{mock_context, noop_assets, MockRuntime};
use tauri_app_lib::create_app_builder;
use tauri_plugin_any_sync::AnySyncExt;

/// Helper function to create a test app instance
fn create_test_app() -> tauri::App<MockRuntime> {
    // Initialize test logging
    let _ = env_logger::builder()
        .is_test(true)
        .filter_level(log::LevelFilter::Debug)
        .try_init();

    // Create the app using the same builder as production but with MockRuntime
    create_app_builder::<MockRuntime>()
        .build(mock_context(noop_assets()))
        .expect("failed to build test app")
}

#[tokio::test]
async fn test_ping_command() {
    let app = create_test_app();

    // Test ping with a message
    let payload = tauri_plugin_any_sync::PingRequest {
        value: Some("Hello from integration test!".to_string()),
    };

    let result = app.any_sync().ping(payload).await;
    assert!(result.is_ok(), "Ping command failed: {:?}", result.err());

    let response = result.unwrap();
    assert!(response.value.is_some(), "Expected a response message");
    assert!(
        response.value.unwrap().contains("Hello from integration test!"),
        "Response should echo the input message"
    );
}

#[tokio::test]
async fn test_ping_command_empty_message() {
    let app = create_test_app();

    // Test ping with no message
    let payload = tauri_plugin_any_sync::PingRequest { value: None };

    let result = app.any_sync().ping(payload).await;
    assert!(
        result.is_ok(),
        "Ping command with empty message failed: {:?}",
        result.err()
    );

    let response = result.unwrap();
    assert!(response.value.is_some(), "Expected a response");
}

#[tokio::test]
async fn test_storage_put_and_get() {
    let app = create_test_app();

    let collection = "test_collection";
    let id = "test_doc_1";
    let document = json!({
        "name": "Test Document",
        "value": 42,
        "nested": {
            "field": "data"
        }
    });

    // Put a document
    let put_payload = tauri_plugin_any_sync::PutRequest {
        collection: collection.to_string(),
        id: id.to_string(),
        document_json: document.to_string(),
    };

    let put_result = app.any_sync().storage_put(put_payload).await;
    assert!(put_result.is_ok(), "Put command failed: {:?}", put_result.err());
    assert!(put_result.unwrap().success, "Put operation should succeed");

    // Get the same document
    let get_payload = tauri_plugin_any_sync::GetRequest {
        collection: collection.to_string(),
        id: id.to_string(),
    };

    let get_result = app.any_sync().storage_get(get_payload).await;
    assert!(get_result.is_ok(), "Get command failed: {:?}", get_result.err());

    let get_response = get_result.unwrap();
    assert!(get_response.found, "Document should be found");
    assert!(
        get_response.document_json.is_some(),
        "Document JSON should be present"
    );

    // Verify the retrieved document matches what we stored
    let retrieved_json = get_response.document_json.unwrap();
    let retrieved: serde_json::Value =
        serde_json::from_str(&retrieved_json).expect("Invalid JSON in response");
    let original: serde_json::Value =
        serde_json::from_str(&document.to_string()).expect("Invalid JSON in original");

    assert_eq!(
        retrieved, original,
        "Retrieved document should match original"
    );
}

#[tokio::test]
async fn test_storage_get_nonexistent() {
    let app = create_test_app();

    // Try to get a document that doesn't exist
    let get_payload = tauri_plugin_any_sync::GetRequest {
        collection: "nonexistent_collection".to_string(),
        id: "nonexistent_id".to_string(),
    };

    let get_result = app.any_sync().storage_get(get_payload).await;
    assert!(get_result.is_ok(), "Get command failed: {:?}", get_result.err());

    let get_response = get_result.unwrap();
    assert!(!get_response.found, "Document should not be found");
    assert!(
        get_response.document_json.is_none(),
        "Document JSON should be None"
    );
}

#[tokio::test]
async fn test_storage_list() {
    let app = create_test_app();

    let collection = "list_test_collection";

    // Put multiple documents
    for i in 1..=5 {
        let put_payload = tauri_plugin_any_sync::PutRequest {
            collection: collection.to_string(),
            id: format!("doc_{}", i),
            document_json: json!({"index": i}).to_string(),
        };

        let result = app.any_sync().storage_put(put_payload).await;
        assert!(result.is_ok(), "Put command {} failed: {:?}", i, result.err());
    }

    // List all documents in the collection
    let list_payload = tauri_plugin_any_sync::ListRequest {
        collection: collection.to_string(),
    };

    let list_result = app.any_sync().storage_list(list_payload).await;
    assert!(
        list_result.is_ok(),
        "List command failed: {:?}",
        list_result.err()
    );

    let list_response = list_result.unwrap();
    assert_eq!(
        list_response.ids.len(),
        5,
        "Should have 5 documents in the collection"
    );

    // Verify all expected IDs are present
    for i in 1..=5 {
        let expected_id = format!("doc_{}", i);
        assert!(
            list_response.ids.contains(&expected_id),
            "Should contain document ID: {}",
            expected_id
        );
    }
}

#[tokio::test]
async fn test_storage_list_empty() {
    let app = create_test_app();

    // List from an empty collection
    let list_payload = tauri_plugin_any_sync::ListRequest {
        collection: "empty_collection".to_string(),
    };

    let list_result = app.any_sync().storage_list(list_payload).await;
    assert!(
        list_result.is_ok(),
        "List command failed: {:?}",
        list_result.err()
    );

    let list_response = list_result.unwrap();
    assert_eq!(
        list_response.ids.len(),
        0,
        "Empty collection should have 0 documents"
    );
}

#[tokio::test]
async fn test_storage_delete() {
    let app = create_test_app();

    let collection = "delete_test_collection";
    let id = "doc_to_delete";

    // Put a document
    let put_payload = tauri_plugin_any_sync::PutRequest {
        collection: collection.to_string(),
        id: id.to_string(),
        document_json: json!({"test": "data"}).to_string(),
    };

    let put_result = app.any_sync().storage_put(put_payload).await;
    assert!(put_result.is_ok(), "Put command failed: {:?}", put_result.err());

    // Delete the document
    let delete_payload = tauri_plugin_any_sync::DeleteRequest {
        collection: collection.to_string(),
        id: id.to_string(),
    };

    let delete_result = app.any_sync().storage_delete(delete_payload).await;
    assert!(
        delete_result.is_ok(),
        "Delete command failed: {:?}",
        delete_result.err()
    );

    let delete_response = delete_result.unwrap();
    assert!(
        delete_response.existed,
        "Document should have existed before deletion"
    );

    // Verify the document is gone
    let get_payload = tauri_plugin_any_sync::GetRequest {
        collection: collection.to_string(),
        id: id.to_string(),
    };

    let get_result = app.any_sync().storage_get(get_payload).await;
    assert!(get_result.is_ok(), "Get command failed: {:?}", get_result.err());

    let get_response = get_result.unwrap();
    assert!(
        !get_response.found,
        "Document should not be found after deletion"
    );
}

#[tokio::test]
async fn test_storage_delete_nonexistent() {
    let app = create_test_app();

    // Try to delete a document that doesn't exist
    let delete_payload = tauri_plugin_any_sync::DeleteRequest {
        collection: "nonexistent_collection".to_string(),
        id: "nonexistent_id".to_string(),
    };

    let delete_result = app.any_sync().storage_delete(delete_payload).await;
    assert!(
        delete_result.is_ok(),
        "Delete command failed: {:?}",
        delete_result.err()
    );

    let delete_response = delete_result.unwrap();
    assert!(
        !delete_response.existed,
        "Nonexistent document should report as not existed"
    );
}

#[tokio::test]
async fn test_storage_update_existing_document() {
    let app = create_test_app();

    let collection = "update_test_collection";
    let id = "doc_to_update";

    // Put initial document
    let put_payload_1 = tauri_plugin_any_sync::PutRequest {
        collection: collection.to_string(),
        id: id.to_string(),
        document_json: json!({"version": 1, "data": "initial"}).to_string(),
    };

    let result = app.any_sync().storage_put(put_payload_1).await;
    assert!(result.is_ok(), "Initial put failed: {:?}", result.err());

    // Update the document
    let put_payload_2 = tauri_plugin_any_sync::PutRequest {
        collection: collection.to_string(),
        id: id.to_string(),
        document_json: json!({"version": 2, "data": "updated"}).to_string(),
    };

    let result = app.any_sync().storage_put(put_payload_2).await;
    assert!(result.is_ok(), "Update put failed: {:?}", result.err());

    // Get and verify the updated document
    let get_payload = tauri_plugin_any_sync::GetRequest {
        collection: collection.to_string(),
        id: id.to_string(),
    };

    let get_result = app.any_sync().storage_get(get_payload).await;
    assert!(get_result.is_ok(), "Get command failed: {:?}", get_result.err());

    let get_response = get_result.unwrap();
    assert!(get_response.found, "Document should be found");

    let retrieved: serde_json::Value =
        serde_json::from_str(&get_response.document_json.unwrap()).expect("Invalid JSON");

    assert_eq!(
        retrieved["version"], 2,
        "Document should have updated version"
    );
    assert_eq!(
        retrieved["data"], "updated",
        "Document should have updated data"
    );
}

#[tokio::test]
async fn test_multiple_collections() {
    let app = create_test_app();

    let collections = vec!["collection_1", "collection_2", "collection_3"];

    // Put documents in different collections
    for collection in &collections {
        let put_payload = tauri_plugin_any_sync::PutRequest {
            collection: collection.to_string(),
            id: "doc_1".to_string(),
            document_json: json!({"collection": collection}).to_string(),
        };

        let result = app.any_sync().storage_put(put_payload).await;
        assert!(
            result.is_ok(),
            "Put to {} failed: {:?}",
            collection,
            result.err()
        );
    }

    // Verify each collection has its document
    for collection in &collections {
        let get_payload = tauri_plugin_any_sync::GetRequest {
            collection: collection.to_string(),
            id: "doc_1".to_string(),
        };

        let get_result = app.any_sync().storage_get(get_payload).await;
        assert!(
            get_result.is_ok(),
            "Get from {} failed: {:?}",
            collection,
            get_result.err()
        );

        let get_response = get_result.unwrap();
        assert!(get_response.found, "Document should be found in {}", collection);

        let retrieved: serde_json::Value =
            serde_json::from_str(&get_response.document_json.unwrap()).expect("Invalid JSON");
        assert_eq!(
            retrieved["collection"], *collection,
            "Document should belong to {}",
            collection
        );
    }
}

#[tokio::test]
async fn test_complex_json_document() {
    let app = create_test_app();

    let collection = "complex_json_collection";
    let id = "complex_doc";

    // Test with a complex nested JSON structure
    let complex_doc = json!({
        "string": "test value",
        "number": 12345,
        "float": 123.456,
        "boolean": true,
        "null_value": null,
        "array": [1, 2, 3, "four", {"nested": "object"}],
        "nested": {
            "level1": {
                "level2": {
                    "level3": "deep value"
                }
            }
        },
        "unicode": "Hello 世界 🌍",
        "special_chars": "quotes:\"' backslash:\\ newline:\n tab:\t"
    });

    let put_payload = tauri_plugin_any_sync::PutRequest {
        collection: collection.to_string(),
        id: id.to_string(),
        document_json: complex_doc.to_string(),
    };

    let put_result = app.any_sync().storage_put(put_payload).await;
    assert!(
        put_result.is_ok(),
        "Put complex document failed: {:?}",
        put_result.err()
    );

    // Retrieve and verify
    let get_payload = tauri_plugin_any_sync::GetRequest {
        collection: collection.to_string(),
        id: id.to_string(),
    };

    let get_result = app.any_sync().storage_get(get_payload).await;
    assert!(get_result.is_ok(), "Get command failed: {:?}", get_result.err());

    let get_response = get_result.unwrap();
    assert!(get_response.found, "Complex document should be found");

    let retrieved: serde_json::Value =
        serde_json::from_str(&get_response.document_json.unwrap()).expect("Invalid JSON");

    // Verify complex structure is preserved
    assert_eq!(retrieved["string"], "test value");
    assert_eq!(retrieved["number"], 12345);
    assert_eq!(retrieved["nested"]["level1"]["level2"]["level3"], "deep value");
    assert_eq!(retrieved["array"][4]["nested"], "object");
    assert_eq!(retrieved["unicode"], "Hello 世界 🌍");
}
