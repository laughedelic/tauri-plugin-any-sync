// Note: Go mobile framework (gomobile-generated) is provided at build time via Package.swift binaryTarget
import AnySync
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
      var error: NSError?
      let success = MobileInit(&error)

      // Check if an error occurred
      if let error = error {
          NSLog("[AnySyncPlugin] MobileInit failed: \(error.localizedDescription)")
          throw error
      }

      // MobileInit returns BOOL indicating success/failure
      if !success {
          let errorMsg = "MobileInit returned false without error details"
          NSLog("[AnySyncPlugin] \(errorMsg)")
          throw NSError(domain: "AnySyncPlugin", code: -1, userInfo: [NSLocalizedDescriptionKey: errorMsg])
      }

      NSLog("[AnySyncPlugin] MobileInit succeeded")
      initialized = true
    }
  }

  @objc public func command(_ invoke: Invoke) throws {
    do {
      try ensureInitialized()

      let args = try invoke.parseArgs(CommandArgs.self)
      let data = Data(args.data)

      // Call Go via gomobile FFI
      var error: NSError?
      let response = MobileCommand(args.cmd, data, &error)

      // Check if an error occurred
      if let error = error {
          NSLog("[AnySyncPlugin] MobileCommand failed for cmd=\(args.cmd): \(error.localizedDescription)")
          throw error
      }

      // Handle nil response from gomobile
      // Note: gomobile converts empty Go slices ([]byte{} with len==0) to nil in Swift/Objective-C
      // This is documented behavior, not a bug. Treat nil as empty response.
      let responseBytes = response != nil ? [UInt8](response!) : []
      NSLog("[AnySyncPlugin] Command '\(args.cmd)' succeeded, response.len=\(responseBytes.count)")
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
