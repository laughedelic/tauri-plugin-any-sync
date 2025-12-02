/// Generic command request/response types for single-dispatch pattern.
/// The actual command structure is handled by protobuf serialization.

/// Command operation - cmd name and opaque request bytes
#[derive(Debug)]
pub struct CommandRequest {
    pub cmd: String,
    pub data: Vec<u8>,
}

/// Command response - opaque response bytes
pub type CommandResponse = Vec<u8>;
