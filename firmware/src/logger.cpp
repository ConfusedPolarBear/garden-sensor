#include <logger.h>

void logInternal(String level, String tag, String message) {
    String color = "";
    String colorReset = "";

    #ifndef LOG_NO_COLOR
    color = "\033[1;";
    colorReset = "\033[0m";

    if (level == "W") {
        color += "33m";     // yellow
    }
    
    else if (level == "F") {
        color += "31m";     // red
    }

    else {
        color = "";
        colorReset = "";
    }
    #endif

    level = "(" + level + ") ";
    Serial << color << level << tag << ": " << message << colorReset << endl;
}

void LOGD(String tag, String message) {
    #ifndef DEBUG
    return;
    #endif

    logInternal("D", tag, message);
}

void LOGW(String tag, String message) {
    logInternal("W", tag, message);
}

void LOGF(String tag, String message) {
    logInternal("F", tag, message);
    ESP.restart();
}