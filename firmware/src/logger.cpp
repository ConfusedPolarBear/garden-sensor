#include <logger.h>

void logInternal(String level, String tag, String message) {
    level = "(" + level + ") ";
    Serial << level << tag << ": " << message << endl;
}

void LOGD(String tag, String message) {
    #ifndef DEBUG
    return;
    #endif

    logInternal("D", tag, message);
}

void LOGF(String tag, String message) {
    logInternal("F", tag, message);
    ESP.restart();
}