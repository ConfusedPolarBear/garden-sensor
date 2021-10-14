#include <logger.h>
#include <LittleFS.h>

#ifdef ESP32
    struct FSInfo {
        size_t usedBytes;
        size_t totalBytes;
    };
#endif

void InitializeFS() {
    LOGD("fs", "mounting filesystem");

    if (!LittleFS.begin()) {
        LOGF("fs", "unable to mount");
    }

    LOGD("fs", "filesystem mounted");
}

void Unmount() {
    LittleFS.end();
}

bool FileExists(String path) {
    return LittleFS.exists(path);
}

void DeleteFile(String path) {
    LittleFS.remove(path);
}

String ReadFile(String path) {
    LOGD("fs", "opening (r) " + path);
    File f = LittleFS.open(path, "r");

    LOGD("fs", "reading contents");
    String contents = f.readString();

    LOGD("fs", "closing");
    f.close();

    LOGD("fs", "returning from read");
    return contents;
}

void WriteFile(String path, String contents) {
    LOGD("fs", "opening (w) " + path);
    File f = LittleFS.open(path, "w");

    LOGD("fs", "writing " + String(contents.length()) + " bytes");
    f.write(contents.c_str(), contents.length());

    LOGD("fs", "closing");
    f.close();

    LOGD("fs", "returning from write");
}

bool IsConfigured() {
    return FileExists("/configured");
}

void SetConfigured(bool flag) {
    if (flag) {
        WriteFile("/configured", "true");
    } else {
        DeleteFile("/configured");
    }
}

bool GetFSInfo(FSInfo* info) {
    #ifdef ESP32
    *info = { LittleFS.usedBytes(), LittleFS.totalBytes() };
    return true;

    #else

    return LittleFS.info(*info);
    #endif
}
