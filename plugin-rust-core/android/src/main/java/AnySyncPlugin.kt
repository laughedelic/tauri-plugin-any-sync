package com.plugin.anysync

import android.app.Activity
import android.util.Log
import app.tauri.annotation.Command
import app.tauri.annotation.InvokeArg
import app.tauri.annotation.TauriPlugin
import app.tauri.plugin.JSObject
import app.tauri.plugin.Plugin
import app.tauri.plugin.Invoke
import mobile.Mobile
import org.json.JSONArray

@InvokeArg
class CommandArgs {
    var cmd: String = ""
    var data: ByteArray = ByteArray(0)
}

@TauriPlugin
class AnySyncPlugin(private val activity: Activity): Plugin(activity) {
    companion object {
        private const val TAG = "AnySync"
        
        init {
            try {
                System.loadLibrary("gojni")
                Log.d(TAG, "Successfully loaded gojni library")
            } catch (e: UnsatisfiedLinkError) {
                Log.e(TAG, "Failed to load gojni library", e)
                throw e
            }
        }
    }

    private var initialized = false

    private fun ensureInitialized() {
        if (!initialized) {
            try {
                Mobile.init()
                initialized = true
                Log.d(TAG, "Mobile backend initialized")
            } catch (e: Exception) {
                Log.e(TAG, "Failed to initialize mobile backend", e)
                throw e
            }
        }
    }

    @Command
    fun command(invoke: Invoke) {
        try {
            ensureInitialized()

            val args = invoke.parseArgs(CommandArgs::class.java)
            Log.d(TAG, "command: cmd=${args.cmd}, data.size=${args.data.size}")

            // Call Go via gomobile FFI
            val response = try {
                Mobile.command(args.cmd, args.data)
            } catch (e: Exception) {
                Log.e(TAG, "Go command threw exception: ${e.message}", e)
                invoke.reject(e.message ?: "Go command failed")
                return
            }

            // Handle null response from gomobile
            // Note: gomobile converts empty Go slices ([]byte{} with len==0) to null in Java/Kotlin
            // This is documented behavior, not a bug. Treat null as empty response.
            val responseBytes = response ?: ByteArray(0)

            Log.d(TAG, "command succeeded, response.size=${responseBytes.size}")

            val ret = JSObject()
            // Convert ByteArray to JSONArray for proper JSON serialization
            val jsonArray = JSONArray()
            responseBytes.forEach { byte -> jsonArray.put(byte.toInt() and 0xFF) }
            ret.put("data", jsonArray)
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "command failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }
}
