import type { DescMessage, Message, MessageInitShape, MessageShape } from "@bufbuild/protobuf";
import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { invoke } from "@tauri-apps/api/core";

import * as pb from "./syncspace_pb";

type Expand<T> = { [K in keyof T]: T[K] } & {};
export type Command = Expand<Omit<pb.Command, keyof Message<"syncspace.v1.Command">>>;

export type CommandResponse = Expand<
  Omit<pb.CommandResponse, keyof Message<"syncspace.v1.CommandResponse">>
>;

export type InitRequest = Expand<Omit<pb.InitRequest, keyof Message<"syncspace.v1.InitRequest">>>;

export type InitResponse = Expand<
  Omit<pb.InitResponse, keyof Message<"syncspace.v1.InitResponse">>
>;

export type ShutdownRequest = Expand<
  Omit<pb.ShutdownRequest, keyof Message<"syncspace.v1.ShutdownRequest">>
>;

export type ShutdownResponse = Expand<
  Omit<pb.ShutdownResponse, keyof Message<"syncspace.v1.ShutdownResponse">>
>;

export type CreateSpaceRequest = Expand<
  Omit<pb.CreateSpaceRequest, keyof Message<"syncspace.v1.CreateSpaceRequest">>
>;

export type CreateSpaceResponse = Expand<
  Omit<pb.CreateSpaceResponse, keyof Message<"syncspace.v1.CreateSpaceResponse">>
>;

export type JoinSpaceRequest = Expand<
  Omit<pb.JoinSpaceRequest, keyof Message<"syncspace.v1.JoinSpaceRequest">>
>;

export type JoinSpaceResponse = Expand<
  Omit<pb.JoinSpaceResponse, keyof Message<"syncspace.v1.JoinSpaceResponse">>
>;

export type LeaveSpaceRequest = Expand<
  Omit<pb.LeaveSpaceRequest, keyof Message<"syncspace.v1.LeaveSpaceRequest">>
>;

export type LeaveSpaceResponse = Expand<
  Omit<pb.LeaveSpaceResponse, keyof Message<"syncspace.v1.LeaveSpaceResponse">>
>;

export type ListSpacesRequest = Expand<
  Omit<pb.ListSpacesRequest, keyof Message<"syncspace.v1.ListSpacesRequest">>
>;

export type ListSpacesResponse = Expand<
  Omit<pb.ListSpacesResponse, keyof Message<"syncspace.v1.ListSpacesResponse">>
>;

export type SpaceInfo = Expand<Omit<pb.SpaceInfo, keyof Message<"syncspace.v1.SpaceInfo">>>;

export type DeleteSpaceRequest = Expand<
  Omit<pb.DeleteSpaceRequest, keyof Message<"syncspace.v1.DeleteSpaceRequest">>
>;

export type DeleteSpaceResponse = Expand<
  Omit<pb.DeleteSpaceResponse, keyof Message<"syncspace.v1.DeleteSpaceResponse">>
>;

export type CreateDocumentRequest = Expand<
  Omit<pb.CreateDocumentRequest, keyof Message<"syncspace.v1.CreateDocumentRequest">>
>;

export type CreateDocumentResponse = Expand<
  Omit<pb.CreateDocumentResponse, keyof Message<"syncspace.v1.CreateDocumentResponse">>
>;

export type GetDocumentRequest = Expand<
  Omit<pb.GetDocumentRequest, keyof Message<"syncspace.v1.GetDocumentRequest">>
>;

export type GetDocumentResponse = Expand<
  Omit<pb.GetDocumentResponse, keyof Message<"syncspace.v1.GetDocumentResponse">>
>;

export type Document = Expand<Omit<pb.Document, keyof Message<"syncspace.v1.Document">>>;

export type UpdateDocumentRequest = Expand<
  Omit<pb.UpdateDocumentRequest, keyof Message<"syncspace.v1.UpdateDocumentRequest">>
>;

export type UpdateDocumentResponse = Expand<
  Omit<pb.UpdateDocumentResponse, keyof Message<"syncspace.v1.UpdateDocumentResponse">>
>;

export type DeleteDocumentRequest = Expand<
  Omit<pb.DeleteDocumentRequest, keyof Message<"syncspace.v1.DeleteDocumentRequest">>
>;

export type DeleteDocumentResponse = Expand<
  Omit<pb.DeleteDocumentResponse, keyof Message<"syncspace.v1.DeleteDocumentResponse">>
>;

export type ListDocumentsRequest = Expand<
  Omit<pb.ListDocumentsRequest, keyof Message<"syncspace.v1.ListDocumentsRequest">>
>;

export type ListDocumentsResponse = Expand<
  Omit<pb.ListDocumentsResponse, keyof Message<"syncspace.v1.ListDocumentsResponse">>
>;

export type DocumentInfo = Expand<
  Omit<pb.DocumentInfo, keyof Message<"syncspace.v1.DocumentInfo">>
>;

export type QueryDocumentsRequest = Expand<
  Omit<pb.QueryDocumentsRequest, keyof Message<"syncspace.v1.QueryDocumentsRequest">>
>;

export type QueryFilter = Expand<Omit<pb.QueryFilter, keyof Message<"syncspace.v1.QueryFilter">>>;

export type QueryDocumentsResponse = Expand<
  Omit<pb.QueryDocumentsResponse, keyof Message<"syncspace.v1.QueryDocumentsResponse">>
>;

export type StartSyncRequest = Expand<
  Omit<pb.StartSyncRequest, keyof Message<"syncspace.v1.StartSyncRequest">>
>;

export type StartSyncResponse = Expand<
  Omit<pb.StartSyncResponse, keyof Message<"syncspace.v1.StartSyncResponse">>
>;

export type PauseSyncRequest = Expand<
  Omit<pb.PauseSyncRequest, keyof Message<"syncspace.v1.PauseSyncRequest">>
>;

export type PauseSyncResponse = Expand<
  Omit<pb.PauseSyncResponse, keyof Message<"syncspace.v1.PauseSyncResponse">>
>;

export type GetSyncStatusRequest = Expand<
  Omit<pb.GetSyncStatusRequest, keyof Message<"syncspace.v1.GetSyncStatusRequest">>
>;

export type GetSyncStatusResponse = Expand<
  Omit<pb.GetSyncStatusResponse, keyof Message<"syncspace.v1.GetSyncStatusResponse">>
>;

export type SpaceSyncStatus = Expand<
  Omit<pb.SpaceSyncStatus, keyof Message<"syncspace.v1.SpaceSyncStatus">>
>;

export type SubscribeRequest = Expand<
  Omit<pb.SubscribeRequest, keyof Message<"syncspace.v1.SubscribeRequest">>
>;

export type SubscribeResponse = Expand<
  Omit<pb.SubscribeResponse, keyof Message<"syncspace.v1.SubscribeResponse">>
>;

export type DocumentCreatedEvent = Expand<
  Omit<pb.DocumentCreatedEvent, keyof Message<"syncspace.v1.DocumentCreatedEvent">>
>;

export type DocumentUpdatedEvent = Expand<
  Omit<pb.DocumentUpdatedEvent, keyof Message<"syncspace.v1.DocumentUpdatedEvent">>
>;

export type DocumentDeletedEvent = Expand<
  Omit<pb.DocumentDeletedEvent, keyof Message<"syncspace.v1.DocumentDeletedEvent">>
>;

export type SyncStatusChangedEvent = Expand<
  Omit<pb.SyncStatusChangedEvent, keyof Message<"syncspace.v1.SyncStatusChangedEvent">>
>;

/**
 * SyncSpaceService provides the complete SyncSpace API for spaces, documents, and synchronization
 * Note: This service definition is for documentation and TypeScript client generation.
 * The actual Go implementation uses the dispatcher pattern with TransportService (see transport.proto).
 *
 * @generated from service syncspace.v1.SyncSpaceService
 */
export class SyncSpaceClient {
  private async dispatch<
    ReqSchema extends DescMessage,
    ReqShape extends MessageInitShape<ReqSchema>,
    ResSchema extends DescMessage,
    ResShape extends MessageShape<ResSchema>,
  >(cmd: string, reqSchema: ReqSchema, resSchema: ResSchema, req: ReqShape): Promise<ResShape> {
    try {
      // Serialize: Interface -> Uint8Array
      const instance = create(reqSchema, req);
      const data = toBinary(reqSchema, instance);

      // Invoke: Uint8Array -> Uint8Array (Direct!)
      const response = await invoke<Uint8Array>("plugin:any-sync|command", { cmd, data });

      // Deserialize: Uint8Array -> Interface
      return fromBinary(resSchema, response) as ResShape;
    } catch (error) {
      throw new Error(
        `Failed to execute command '${cmd}': ${error instanceof Error ? error.message : String(error)}`,
      );
    }
  }

  /**
   * Lifecycle operations
   *
   * @generated from rpc syncspace.v1.SyncSpaceService.Init
   */
  public async init(request: InitRequest): Promise<InitResponse> {
    return await this.dispatch("Init", pb.InitRequestSchema, pb.InitResponseSchema, request);
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.Shutdown
   */
  public async shutdown(): Promise<ShutdownResponse> {
    return await this.dispatch("Shutdown", pb.ShutdownRequestSchema, pb.ShutdownResponseSchema, {});
  }

  /**
   * Space operations
   *
   * @generated from rpc syncspace.v1.SyncSpaceService.CreateSpace
   */
  public async createSpace(request: CreateSpaceRequest): Promise<CreateSpaceResponse> {
    return await this.dispatch(
      "CreateSpace",
      pb.CreateSpaceRequestSchema,
      pb.CreateSpaceResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.JoinSpace
   */
  public async joinSpace(request: JoinSpaceRequest): Promise<JoinSpaceResponse> {
    return await this.dispatch(
      "JoinSpace",
      pb.JoinSpaceRequestSchema,
      pb.JoinSpaceResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.LeaveSpace
   */
  public async leaveSpace(request: LeaveSpaceRequest): Promise<LeaveSpaceResponse> {
    return await this.dispatch(
      "LeaveSpace",
      pb.LeaveSpaceRequestSchema,
      pb.LeaveSpaceResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.ListSpaces
   */
  public async listSpaces(): Promise<ListSpacesResponse> {
    return await this.dispatch(
      "ListSpaces",
      pb.ListSpacesRequestSchema,
      pb.ListSpacesResponseSchema,
      {},
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.DeleteSpace
   */
  public async deleteSpace(request: DeleteSpaceRequest): Promise<DeleteSpaceResponse> {
    return await this.dispatch(
      "DeleteSpace",
      pb.DeleteSpaceRequestSchema,
      pb.DeleteSpaceResponseSchema,
      request,
    );
  }

  /**
   * Document operations
   *
   * @generated from rpc syncspace.v1.SyncSpaceService.CreateDocument
   */
  public async createDocument(request: CreateDocumentRequest): Promise<CreateDocumentResponse> {
    return await this.dispatch(
      "CreateDocument",
      pb.CreateDocumentRequestSchema,
      pb.CreateDocumentResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.GetDocument
   */
  public async getDocument(request: GetDocumentRequest): Promise<GetDocumentResponse> {
    return await this.dispatch(
      "GetDocument",
      pb.GetDocumentRequestSchema,
      pb.GetDocumentResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.UpdateDocument
   */
  public async updateDocument(request: UpdateDocumentRequest): Promise<UpdateDocumentResponse> {
    return await this.dispatch(
      "UpdateDocument",
      pb.UpdateDocumentRequestSchema,
      pb.UpdateDocumentResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.DeleteDocument
   */
  public async deleteDocument(request: DeleteDocumentRequest): Promise<DeleteDocumentResponse> {
    return await this.dispatch(
      "DeleteDocument",
      pb.DeleteDocumentRequestSchema,
      pb.DeleteDocumentResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.ListDocuments
   */
  public async listDocuments(request: ListDocumentsRequest): Promise<ListDocumentsResponse> {
    return await this.dispatch(
      "ListDocuments",
      pb.ListDocumentsRequestSchema,
      pb.ListDocumentsResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.QueryDocuments
   */
  public async queryDocuments(request: QueryDocumentsRequest): Promise<QueryDocumentsResponse> {
    return await this.dispatch(
      "QueryDocuments",
      pb.QueryDocumentsRequestSchema,
      pb.QueryDocumentsResponseSchema,
      request,
    );
  }

  /**
   * Sync control operations
   *
   * @generated from rpc syncspace.v1.SyncSpaceService.StartSync
   */
  public async startSync(request: StartSyncRequest): Promise<StartSyncResponse> {
    return await this.dispatch(
      "StartSync",
      pb.StartSyncRequestSchema,
      pb.StartSyncResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.PauseSync
   */
  public async pauseSync(request: PauseSyncRequest): Promise<PauseSyncResponse> {
    return await this.dispatch(
      "PauseSync",
      pb.PauseSyncRequestSchema,
      pb.PauseSyncResponseSchema,
      request,
    );
  }

  /**
   * @generated from rpc syncspace.v1.SyncSpaceService.GetSyncStatus
   */
  public async getSyncStatus(request: GetSyncStatusRequest): Promise<GetSyncStatusResponse> {
    return await this.dispatch(
      "GetSyncStatus",
      pb.GetSyncStatusRequestSchema,
      pb.GetSyncStatusResponseSchema,
      request,
    );
  }

  /**
   * Event streaming
   *
   * @generated from rpc syncspace.v1.SyncSpaceService.Subscribe
   */
  public async subscribe(request: SubscribeRequest): Promise<SubscribeResponse> {
    return await this.dispatch(
      "Subscribe",
      pb.SubscribeRequestSchema,
      pb.SubscribeResponseSchema,
      request,
    );
  }
}

export const syncspace = new SyncSpaceClient();
