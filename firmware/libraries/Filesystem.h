#pragma once

#include <LittleFS.h>

class Filesystem {
    private:
        static Filesystem* _instance;

        Filesystem();

    public:
        // Prohibit cloning and assigning the filesystem object.
        Filesystem(Filesystem &other) = delete;
        void operator=(const Filesystem &) = delete;

        // Returns the filesystem, opening it first if necessary.
        static Filesystem* Get();

        // Unmount the filesystem. No file operations should be performed after this!
        static void Unmount();

        static bool FileExists(String path);

        static String ReadFile(String path);
        static void WriteFile(String path, String contents);
        static bool DeleteFile(String path);

        static bool IsConfigured();
        static void SetConfigured(bool flag);

        static bool Format();
};

Filesystem* Filesystem::_instance = 0;

Filesystem::Filesystem() {
    if (!LittleFS.begin()) {
        Serial << "filesystem fatal error: unable to mount internal filesystem" << endl;

        while(true) {
            yield();
        }
    }
}

Filesystem* Filesystem::Get() {
    if (_instance == nullptr) { _instance = new Filesystem(); }
    return _instance;
}

void Filesystem::Unmount() {
    LittleFS.end();
    delete _instance;
}

bool Filesystem::FileExists(String path) {
    return LittleFS.exists(path);
}

String Filesystem::ReadFile(String path) {
    if (!_instance->FileExists(path)) {
        return "";
    }

    File f = LittleFS.open(path, "r");
    String contents = f.readString();
    f.close();

    return contents;
}

void Filesystem::WriteFile(String path, String contents) {
    File f = LittleFS.open(path, "w");
    f << contents;
    f.close();
}

bool Filesystem::DeleteFile(String path) {
    LittleFS.remove(path);
}

bool Filesystem::IsConfigured() {
    return _instance->FileExists("/configured");
}

void Filesystem::SetConfigured(bool flag) {
    if (flag) {
        _instance->WriteFile("/configured", "true");
    } else {
        _instance->DeleteFile("/configured");
    }
}

bool Filesystem::Format() {
    return LittleFS.format();
}
