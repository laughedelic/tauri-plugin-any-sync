# Storage API Specification

## ADDED Requirements

### Requirement: Storage Service Definition

The Go backend SHALL provide a gRPC StorageService with four CRUD operations for document storage and retrieval.

#### Scenario: Put operation stores document

- **GIVEN** a collection name, document ID, and valid JSON document
- **WHEN** Put RPC is called
- **THEN** the document is stored in AnyStore and success is returned

#### Scenario: Get operation retrieves document

- **GIVEN** a collection name and document ID of an existing document
- **WHEN** Get RPC is called
- **THEN** the document JSON is returned

#### Scenario: Get operation handles missing document

- **GIVEN** a collection name and document ID of a non-existent document
- **WHEN** Get RPC is called
- **THEN** an empty response is returned with found=false

#### Scenario: Delete operation removes document

- **GIVEN** a collection name and document ID of an existing document
- **WHEN** Delete RPC is called
- **THEN** the document is removed from AnyStore and existed=true is returned

#### Scenario: Delete operation is idempotent

- **GIVEN** a collection name and document ID of a non-existent document
- **WHEN** Delete RPC is called
- **THEN** success is returned with existed=false (no error)

#### Scenario: List operation returns all IDs

- **GIVEN** a collection with multiple documents
- **WHEN** List RPC is called
- **THEN** all document IDs in the collection are returned

### Requirement: AnyStore Integration Layer

The Go backend SHALL provide a storage wrapper in `internal/storage/anystore.go` that abstracts AnyStore-specific types.

#### Scenario: Wrapper initializes AnyStore database

- **GIVEN** the sidecar server is starting
- **WHEN** the storage wrapper is initialized
- **THEN** an AnyStore database is opened in the app data directory

#### Scenario: Wrapper converts between gRPC and AnyStore types

- **GIVEN** a gRPC Put request with JSON string
- **WHEN** the wrapper processes the request
- **THEN** the JSON is parsed to `anyenc.Value` and stored via AnyStore

#### Scenario: Wrapper handles AnyStore errors

- **GIVEN** AnyStore returns an error
- **WHEN** the wrapper processes the error
- **THEN** the error is converted to appropriate gRPC status code

### Requirement: Protobuf Message Definitions

The storage service SHALL define gomobile-compatible protobuf messages using only primitive types.

#### Scenario: PutRequest contains required fields

- **GIVEN** a storage Put operation
- **WHEN** the PutRequest message is constructed
- **THEN** it contains collection (string), id (string), and document_json (string)

#### Scenario: GetRequest contains required fields

- **GIVEN** a storage Get operation
- **WHEN** the GetRequest message is constructed
- **THEN** it contains collection (string) and id (string)

#### Scenario: DeleteRequest contains required fields

- **GIVEN** a storage Delete operation
- **WHEN** the DeleteRequest message is constructed
- **THEN** it contains collection (string) and id (string)

#### Scenario: DeleteResponse indicates if document existed

- **GIVEN** a Delete operation completes successfully
- **WHEN** the DeleteResponse is returned
- **THEN** it contains existed (bool) indicating if the document was present

#### Scenario: ListRequest contains required fields

- **GIVEN** a storage List operation
- **WHEN** the ListRequest message is constructed
- **THEN** it contains collection (string)

### Requirement: JSON Document Validation

The storage service SHALL validate JSON documents before storing them in AnyStore.

#### Scenario: Valid JSON is accepted

- **GIVEN** a Put request with well-formed JSON
- **WHEN** validation is performed
- **THEN** the document is stored without error

#### Scenario: Invalid JSON is rejected

- **GIVEN** a Put request with malformed JSON
- **WHEN** validation is performed
- **THEN** an INVALID_ARGUMENT error is returned with details

### Requirement: Collection-Based Organization

The storage service SHALL organize documents into named collections matching AnyStore's collection API.

#### Scenario: Documents are isolated by collection

- **GIVEN** two documents with the same ID in different collections
- **WHEN** both are stored and retrieved
- **THEN** each document is retrieved from its respective collection

#### Scenario: Collections are created automatically

- **GIVEN** a Put request for a non-existent collection
- **WHEN** the document is stored
- **THEN** the collection is created automatically by AnyStore

### Requirement: Error Context Propagation

The storage service SHALL include collection name and document ID in error messages for debugging.

#### Scenario: Get response includes found status

- **GIVEN** a Get request for any document
- **WHEN** the response is returned
- **THEN** the found field indicates whether the document exists

#### Scenario: Invalid argument error includes context

- **GIVEN** a Put request with invalid JSON
- **WHEN** the error is returned
- **THEN** the error message includes collection name and JSON parsing details
