#!/usr/bin/env -S bun run

/**
 * Buf plugin to generate a typed SyncSpace client with generic dispatch pattern.
 * This script uses @bufbuild/protoplugin to integrate with buf's code generation pipeline.
 *
 * The generated client:
 * - Provides one async method per RPC service method
 * - Uses a private generic `call<Req, Res>()` method to reduce duplication
 * - Handles encoding/decoding with protobuf-es functions
 * - Dispatches to the Tauri backend via the `plugin:any-sync|command` invoke
 */

import type { DescMethod, DescService } from "@bufbuild/protobuf";
import {
  createEcmaScriptPlugin,
  type GeneratedFile,
  runNodeJs,
  type Schema,
} from "@bufbuild/protoplugin";

runNodeJs(
  createEcmaScriptPlugin({
    name: "custom-client",
    version: "v1",
    generateTs,
  }),
);

function generateTs(schema: Schema) {
  for (const file of schema.files) {
    // Only generate client for files with services
    if (file.services.length === 0) continue;

    // Create a file generator for *_client.ts
    const f = schema.generateFile(`${file.name}_client.ts`);

    // Add preamble (license, generator info, etc.)
    f.preamble(file);

    // Generate client class for each service
    for (const service of file.services) {
      const MessageInitShape = f.import("MessageInitShape", "@bufbuild/protobuf").toTypeOnly();
      const printedAliases = new Set<string>();

      for (const method of service.methods) {
        const inputType = f.importShape(method.input);
        const inputSchema = f.importSchema(method.input);
        const outputType = f.importShape(method.output);
        const outputSchema = f.importSchema(method.output);
        // Generate an alias for the Input type if we haven't yet
        if (!printedAliases.has(method.input.name)) {
          f.print(
            f.export("type", inputType.name),
            " = ",
            MessageInitShape,
            "<typeof ",
            inputSchema,
            ">;",
          );
          f.print(
            f.export("type", outputType.name),
            " = ",
            MessageInitShape,
            "<typeof ",
            outputSchema,
            ">;",
          );

          printedAliases.add(inputType.name);
          printedAliases.add(outputType.name);
        }
      }

      generateServiceClient(f, service);
    }
  }
}

function generateServiceClient(f: GeneratedFile, service: DescService) {
  const className = service.name;

  // Add JSDoc for the class
  f.print(f.jsDoc(service, "  "));

  // Generate class declaration
  f.print(f.export("class", className), " {");

  generateDispatchCommand(f);
  generateCommandInvoker(f);

  // Generate one method per RPC
  service.methods.forEach((method) => {
    generateServiceMethod(f, method);
  });

  f.print("}");
  f.print();

  // Create singleton export
  // Remove "Service" suffix if present, then lowercase
  let singletonName = service.name.replace(/Service$/, "");
  singletonName = singletonName.charAt(0).toLowerCase() + singletonName.slice(1);

  f.print("// Convenience export for singleton");
  f.print(f.export("const", singletonName), " = new ", className, "();");
}

function generateDispatchCommand(f: GeneratedFile) {
  // Import Tauri invoke function
  const invoke = f.import("invoke", "@tauri-apps/api/core");

  f.print`
    /**
     * Raw command function for dispatching to the SyncSpace backend.
     * Use this for advanced cases where the typed client is insufficient.
     */
    async dispatchCommand(cmd: string, payload: Uint8Array): Promise<Uint8Array> {
      try {
        const responseArray = await ${invoke}<number[]>(
        "plugin:any-sync|command",
        { cmd, payload: Array.from(payload) },
        );

        return new Uint8Array(responseArray);
      } catch (error) {
        throw new Error(
          \`Failed to execute command '\${cmd}': \${error instanceof Error ? error.message : String(error)}\`,
        );
      }
    }
  `;
}

function generateCommandInvoker(f: GeneratedFile) {
  const Message = f.import("Message", "@bufbuild/protobuf").toTypeOnly();
  const GenMessage = f.import("GenMessage", "@bufbuild/protobuf/codegenv2").toTypeOnly();
  const { fromBinary, toBinary } = f.runtime;

  // Private generic call method
  f.print`
    private async _dispatch<Req extends ${Message}, Res extends ${Message}>(
      cmd: string,
      reqSchema: ${GenMessage}<Req>,
      resSchema: ${GenMessage}<Res>,
      req: Req,
    ): Promise<Res> {
      return ${fromBinary}(resSchema, await this.dispatchCommand(cmd, ${toBinary}(reqSchema, req)));
    }
  `;
}

function generateServiceMethod(f: GeneratedFile, method: DescMethod) {
  // Import the request and response types and schemas
  const inputType = f.importShape(method.input);
  const inputSchema = f.importSchema(method.input);
  const outputType = f.importShape(method.output);
  const outputSchema = f.importSchema(method.output);
  const { create } = f.runtime;

  // Add JSDoc for the method
  f.print(f.jsDoc(method, "  "));

  // Generate the method with lowercase first letter
  f.print`
    async ${method.localName}(request: ${inputType.name}): Promise<${outputType.name}> {
      const req = ${create}(${inputSchema}, request);
      return this._dispatch("${method.name}", ${inputSchema}, ${outputSchema}, req);
    }
  `;
}
