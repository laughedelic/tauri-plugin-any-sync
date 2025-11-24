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

@InvokeArg
class StorageGetArgs {
    var collection: String = ""
    var id: String = ""
}

@InvokeArg
class StoragePutArgs {
    var collection: String = ""
    var id: String = ""
    var documentJson: String = ""
}

@InvokeArg
class StorageDeleteArgs {
    var collection: String = ""
    var id: String = ""
}

@InvokeArg
class StorageListArgs {
    var collection: String = ""
}

@TauriPlugin
class ExamplePlugin(private val activity: Activity): Plugin(activity) {
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
            val dbPath = activity.filesDir.absolutePath + "/anysync.db"
            try {
                Mobile.initStorage(dbPath)
                initialized = true
                Log.d(TAG, "Storage initialized at: $dbPath")
            } catch (e: Exception) {
                Log.e(TAG, "Failed to initialize storage", e)
                throw e
            }
        }
    }

    @Command
    fun storageGet(invoke: Invoke) {
        try {
            ensureInitialized()
            val args = invoke.parseArgs(StorageGetArgs::class.java)
            
            Log.d(TAG, "storageGet: collection=${args.collection}, id=${args.id}")
            
            val result = Mobile.storageGet(args.collection, args.id)
            val ret = JSObject()
            // Match GetResponse: documentJson and found fields
            if (result.isNullOrEmpty()) {
                ret.put("documentJson", null)
                ret.put("found", false)
            } else {
                ret.put("documentJson", result)
                ret.put("found", true)
            }
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "storageGet failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }

    @Command
    fun storagePut(invoke: Invoke) {
        try {
            ensureInitialized()
            val args = invoke.parseArgs(StoragePutArgs::class.java)
            
            Log.d(TAG, "storagePut: collection=${args.collection}, id=${args.id}")
            
            Mobile.storagePut(args.collection, args.id, args.documentJson)
            val ret = JSObject()
            ret.put("success", true)
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "storagePut failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }

    @Command
    fun storageDelete(invoke: Invoke) {
        try {
            ensureInitialized()
            val args = invoke.parseArgs(StorageDeleteArgs::class.java)
            
            Log.d(TAG, "storageDelete: collection=${args.collection}, id=${args.id}")
            
            val existed = Mobile.storageDelete(args.collection, args.id)
            val ret = JSObject()
            // Match DeleteResponse: existed field
            ret.put("existed", existed)
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "storageDelete failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }

    @Command
    fun storageList(invoke: Invoke) {
        try {
            ensureInitialized()
            val args = invoke.parseArgs(StorageListArgs::class.java)
            
            Log.d(TAG, "storageList: collection=${args.collection}")
            
            // Go backend returns JSON array string like ["id1","id2"]
            val result = Mobile.storageList(args.collection)
            val ret = JSObject()
            
            // Handle empty/null results gracefully
            if (result.isNullOrEmpty()) {
                Log.d(TAG, "storageList returned empty/null, using empty array")
                ret.put("ids", org.json.JSONArray())
            } else {
                try {
                    ret.put("ids", org.json.JSONArray(result))
                } catch (e: org.json.JSONException) {
                    Log.w(TAG, "Failed to parse list result as JSON array: $result", e)
                    // Return empty array on parse error
                    ret.put("ids", org.json.JSONArray())
                }
            }
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "storageList failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }
}
