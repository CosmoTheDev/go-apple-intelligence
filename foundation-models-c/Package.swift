/*
For licensing see accompanying LICENSE file.
Copyright (C) 2026 Apple Inc. All Rights Reserved.
*/

// swift-tools-version: 6.2
import PackageDescription

let package = Package(
  name: "foundation-models-c-bindings",
  platforms: [.macOS(.v26), .iOS(.v26), .visionOS(.v26)],
  products: [
    .library(name: "FoundationModels", type: .dynamic, targets: ["FoundationModelsCBindings"]),
    .library(name: "FoundationModelsStatic", type: .static, targets: ["FoundationModelsCBindings"]),
  ],
  targets: [
    .target(
      name: "FoundationModelsCDeclarations"
    ),
    .target(
      name: "FoundationModelsCBindings",
      dependencies: ["FoundationModelsCDeclarations"],
      publicHeadersPath: "include",
      cSettings: [
        .headerSearchPath("Sources/FoundationModelsCBindings/include")
      ]
    ),
  ],
  cLanguageStandard: .c99
)
