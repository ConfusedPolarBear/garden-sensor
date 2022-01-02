#include <logger.h>

void logAdvanced(String level, String tag, String message) {
    String color = "";
    String colorReset = "";

    #ifndef LOG_NO_COLOR
    color = "\033[1;";
    colorReset = "\033[0m";

    if (level == "I") {
        color += "36m";     // blue
    }

    else if (level == "W") {
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

void LOGI(String tag, String message) {
    logAdvanced("I", tag, message);
}

void LOGW(String tag, String message) {
    logAdvanced("W", tag, message);
}

void LOGF(String tag, String message) {
    logAdvanced("F", tag, message);
    ESP.restart();
    while(1) { yield(); }
}