// Note: Mobile framework (gomobile-generated) is provided at app build time
// Standalone swift build will fail - this is expected and OK
// The framework is located at: ../../binaries/any-sync-ios.xcframework
import Mobile
import SwiftRs
import Tauri
import UIKit
import WebKit

class CommandArgs: Decodable {
  let cmd: String
  let data: [UInt8]
}

class AnySyncPlugin: Plugin {
  private var initialized = false

  private func ensureInitialized() throws {
    if !initialized {
      try MobileInit()
      initialized = true
    }
  }

  @objc public func command(_ invoke: Invoke) throws {
    do {
      try ensureInitialized()

      let args = try invoke.parseArgs(CommandArgs.self)
      let data = Data(args.data)

      // Call Go via gomobile FFI
      let response = try MobileCommand(args.cmd, data)

      // Convert response Data back to byte array for Rust
      let responseBytes = [UInt8](response ?? Data())
      invoke.resolve(["data": responseBytes])

    } catch {
      invoke.reject(error.localizedDescription)
    }
  }
}

@_cdecl("init_plugin_any_sync")
func initPlugin() -> Plugin {
  return AnySyncPlugin()
}
