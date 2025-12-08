// swift-tools-version:5.3
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "tauri-plugin-any-sync",
    platforms: [
        .macOS(.v10_13),
        .iOS(.v13),
    ],
    products: [
        .library(
            name: "tauri-plugin-any-sync",
            type: .static,
            targets: ["tauri-plugin-any-sync"])
    ],
    dependencies: [
        // Tauri dependency is resolved by the consuming app at build time
        .package(name: "Tauri", path: "../.tauri/tauri-api")
    ],
    targets: [
        .target(
            name: "tauri-plugin-any-sync",
            dependencies: [
                .byName(name: "Tauri"),
                "AnySync"
            ],
            path: "Sources"
        ),
        // Go mobile framework (gomobile-generated, copied by build.rs)
        .binaryTarget(
            name: "AnySync",
            path: "../../binaries/AnySync.xcframework"
        ),
        .testTarget(
            name: "tauri-plugin-any-sync-tests",
            dependencies: ["tauri-plugin-any-sync"],
            path: "Tests/PluginTests"),
    ]
)
