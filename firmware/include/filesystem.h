#pragma once

#include <logger.h>
#include <LittleFS.h>

#ifdef ESP32
    // The FSInfo struct doesn't exist in the esp32 core, but we need it
    struct FSInfo {
        size_t usedBytes;
        size_t totalBytes;
    };
#endif

// Mounts the filesystem or panics.
void InitializeFS();

// Unmounts the filesystem.
void Unmount();

// Returns if the specified file exists or not.
bool FileExists(String path);

// Deletes the specified file.
void DeleteFile(String path);

// Returns the contents of the file or the empty string if it does not exist.
String ReadFile(String path);

// Writes the provided string to the named path.
void WriteFile(String path, String contents);

// Checks if this system is configured.
bool IsConfigured();

// Sets the configuration state of this system.
void SetConfigured(bool flag);

// Returns the total and available bytes in the filesystem.
bool GetFSInfo(FSInfo* info);
