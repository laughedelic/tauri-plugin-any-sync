// Note: Go mobile framework (gomobile-generated) is provided at build time via Package.swift binaryTarget
// The framework is symlinked to Frameworks/any-sync-ios.xcframework by build.rs
import Any_Sync_Ios
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
