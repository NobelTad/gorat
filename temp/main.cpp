#include <windows.h>
#include <iostream>
#include <string>
#include <shlwapi.h> // for PathFileExists

#pragma comment(lib, "Shlwapi.lib")

bool create_folder(const char* folderPath) {
    if (CreateDirectory(folderPath, NULL) || GetLastError() == ERROR_ALREADY_EXISTS) {
        if (SetFileAttributes(folderPath, FILE_ATTRIBUTE_HIDDEN)) {
            std::cout << "[+] Folder created and hidden: " << folderPath << "\n";
            return true;
        } else {
            std::cerr << "[-] Failed to hide folder: " << folderPath << "\n";
        }
    } else {
        std::cerr << "[-] Failed to create folder: " << folderPath << "\n";
    }
    return false;
}

bool move_files_to_folder(const char* folderPath, int count, const char* srcFiles[], const char* destNames[]) {
    for (int i = 0; i < count; i++) {
        std::string destPath = std::string(folderPath) + "\\" + destNames[i];

        if (!PathFileExistsA(srcFiles[i])) {
            std::cerr << "[-] Source file not found: " << srcFiles[i] << "\n";
            continue;
        }

        if (MoveFileA(srcFiles[i], destPath.c_str())) {
            std::cout << "[+] Moved: " << srcFiles[i] << " -> " << destPath << "\n";
        } else {
            std::cerr << "[-] Failed to move: " << srcFiles[i] << "\n";
        }
    }
    return true;
}

int main() {
    const char* folder = "C:\\fake\\LINA\\Desktop\\games\\sys64";
    create_folder(folder);

    const char* sourceFiles[] = {
        "test1.txt",
        "test2.txt"
    };

    const char* destNames[] = {
        "doc1.txt",
        "doc2.txt"
    };

    move_files_to_folder(folder, 2, sourceFiles, destNames);

    return 0;
}
