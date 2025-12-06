// Example usage of the generated client
import { syncspace } from ".";

syncspace.createSpace({
  spaceId: "my-space",
  name: "My Space",
  metadata: {},
});

syncspace.joinSpace({
  spaceId: "my-space",
  inviteToken: "",
});

syncspace.listSpaces();
