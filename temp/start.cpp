#include <windows.h>
#include <iostream>
#include <fstream>
#include <string>

int main() {
    // 1) Get %APPDATA%
    char appdata[MAX_PATH];
    if (!GetEnvironmentVariableA("APPDATA", appdata, MAX_PATH)) {
        std::cerr << "[-] Failed to read APPDATA\n";
        return 1;
    }

    // 2) Build Startup folder path
    std::string startup = std::string(appdata)
        + "\\Microsoft\\Windows\\Start Menu\\Programs\\Startup";

    // 3) Determine full path to main.exe
    char exePath[MAX_PATH];
    if (!GetModuleFileNameA(NULL, exePath, MAX_PATH)) {
        std::cerr << "[-] Failed to get module path\n";
        return 1;
    }

    // 4) Name the .bat stub
    std::string batPath = startup + "\\run_main.bat";

    // 5) Write the batch file
    std::ofstream bat(batPath, std::ios::trunc);
    if (!bat) {
        std::cerr << "[-] Cannot open " << batPath << "\n";
        return 1;
    }

    // @echo off: silence the console
    // start "" "C:\full\path\to\main.exe"
    bat << "@echo off\r\n"
        << "start \"\" \"" << exePath << "\"\r\n";
    bat.close();

    std::cout << "[+] Created startup stub at:\n    " << batPath << "\n";
    return 0;
}
