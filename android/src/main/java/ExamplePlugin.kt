package com.plugin.any-sync

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
    var document: String = ""
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
            ret.put("document", result)
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
            
            Mobile.storagePut(args.collection, args.id, args.document)
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
            
            val deleted = Mobile.storageDelete(args.collection, args.id)
            val ret = JSObject()
            ret.put("deleted", deleted)
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
            
            val result = Mobile.storageList(args.collection)
            val ret = JSObject()
            ret.put("documents", result)
            invoke.resolve(ret)
        } catch (e: Exception) {
            Log.e(TAG, "storageList failed", e)
            invoke.reject(e.message ?: "Unknown error")
        }
    }
}
