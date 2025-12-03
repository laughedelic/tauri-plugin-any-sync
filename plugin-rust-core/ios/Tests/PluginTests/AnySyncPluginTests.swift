import XCTest

@testable import ExamplePlugin

final class AnySyncPluginTests: XCTestCase {
    func testPluginInitialization() throws {
        // Test that plugin can be instantiated
        let plugin = AnySyncPlugin()
        XCTAssertNotNil(plugin)
    }

    // Note: Testing command execution requires gomobile framework to be linked
    // and actual Go backend initialization, which is covered by integration tests.
}
