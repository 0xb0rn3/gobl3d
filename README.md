# gobl3d - Ultimate Enumeration !

## Installation

### Quick Setup
```bash
# Clone or download the repository
git clone https://github.com/secvulnhub/gobl3d
cd gobl3d

# Make the wrapper executable
chmod +x gobl3d.sh

# Run the wrapper (it will handle all dependencies)
./gobl3d.sh
```

### Manual Setup
```bash
# Install dependencies
sudo apt-get update
sudo apt-get install golang-go gobuster jq parallel curl wget

# Or for other distributions:
# yum install golang gobuster jq parallel curl wget
# pacman -S go gobuster jq parallel curl wget
# brew install go gobuster jq parallel curl wget

# Make scripts executable
chmod +x gobl3d.sh module.bash
```

## File Structure
```
gobl3d/
├── gobl3d.sh          # Main wrapper script
├── module.go          # Go enumeration module
├── module.bash        # Bash enumeration module
├── config.json        # Configuration template
├── integrate.sh       # Integration helper (auto-generated)
└── README.md          # This file
```

## Configuration Template (config.json)

```json
{
  "target": "https://your-target.com",
  "wordlists": [
    "/usr/share/wordlists/dirb/common.txt",
    "/opt/SecLists/Discovery/Web-Content/directory-list-2.3-medium.txt",
    "/usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt"
  ],
  "extensions": ["php", "html", "txt", "js", "asp", "aspx", "jsp", "zip", "bak"],
  "threads": 30,
  "delay": "100ms",
  "timeout": "10s",
  "output_dir": "./gobl3d_results",
  "skip_ssl": true,
  "follow_redirects": true,
  "user_agent": "Mozilla/5.0 (compatible; gobl3d/1.0)",
  "status_codes": ["200", "301", "302", "401", "403"],
  "exclude_codes": ["404", "500"]
}
```

## Usage Examples

### Basic Usage
```bash
# Run the interactive wrapper
./gobl3d.sh

# The tool will present a tactical menu:
# [1] Quick Recon     │ Fast discovery
# [2] Comprehensive   │ Deep enumeration
# [3] Config Setup    │ Modify settings
# [4] Integration     │ Tool chaining
# [5] Exit            │ Quit program
```

### Menu Options Explained

#### Option 1: Quick Recon
- Fast directory enumeration only
- Limited extensions (php, html, txt, js)
- 20 threads, 5s timeout
- No recursive scanning
- Perfect for initial reconnaissance

#### Option 2: Comprehensive Scan
- Full enumeration suite
- Directory + Subdomain + VHost scanning
- Recursive enumeration on interesting directories
- Uses both Go and Bash modules
- Configurable via config.json

#### Option 3: Config Setup
- Create/modify configuration file
- Interactive target selection
- Advanced parameter tuning
- Template-based configuration

#### Option 4: Integration
- Shows integration examples
- Creates helper scripts
- Tool chaining guidance
- Export formats for other tools

## Integration Examples

### With Burp Suite
```bash
# The tool will automatically detect proxy settings
# Just configure Burp to listen on 127.0.0.1:8080
```

### With ffuf
```bash
# After running gobl3d, use the integration helper:
./integrate.sh ffuf ./gobl3d_results

# This creates ffuf_dirs.txt with discovered directories
ffuf -u https://target.com/FUZZ -w ffuf_dirs.txt
```

### With Nuclei
```bash
# Extract subdomains for Nuclei scanning:
./integrate.sh nuclei ./gobl3d_results

# Run Nuclei with discovered subdomains:
nuclei -l nuclei_targets.txt -t /path/to/nuclei-templates/
```

### With Nmap
```bash
# Prepare targets for Nmap:
./integrate.sh nmap ./gobl3d_results

# Run port scanning:
nmap -iL nmap_targets.txt -p- --min-rate 1000
```

## Performance Tuning

### High-Performance Scan
```json
{
  "threads": 100,
  "timeout": "5s",
  "delay": "50ms"
}
```

### Stealth Scan
```json
{
  "threads": 5,
  "timeout": "30s",
  "delay": "2s"
}
```

### Large Target Scan
```json
{
  "extensions": ["php","asp","aspx","jsp","html","txt","js","css","xml","json","zip","tar","gz","bak","old","tmp"],
  "wordlists": [
    "/opt/SecLists/Discovery/Web-Content/directory-list-2.3-big.txt",
    "/opt/SecLists/Discovery/Web-Content/raft-large-directories.txt"
  ]
}
```

## Advanced Features

### Custom Wordlists
Edit `config.json` and add your wordlist paths:
```json
{
  "wordlists": [
    "/custom/path/to/wordlist.txt",
    "/another/custom/wordlist.txt"
  ]
}
```

### Custom Headers
```json
{
  "headers": [
    "Authorization: Bearer token123",
    "X-Forwarded-For: 127.0.0.1"
  ]
}
```

### Proxy Configuration
```json
{
  "proxy": "http://127.0.0.1:8080"
}
```

## Output Structure

### Quick Recon Output
```
gobl3d_quick_20231120_143022/
├── directories_1.txt
├── summary.txt
└── scan_config.txt
```

### Comprehensive Scan Output
```
gobl3d_results/
├── Go Module Results/
│   ├── gobuster_scan_20231120_143022/
│   ├── directories_1.txt
│   ├── subdomains_1.txt
│   ├── vhosts_1.txt
│   └── report.json
├── Bash Module Results/
│   ├── gobuster_results_20231120_143022_bash/
│   ├── directories_1.txt
│   ├── recursive_admin.txt
│   └── summary.txt
└── Combined Results/
    ├── all_directories.txt
    ├── all_subdomains.txt
    └── final_report.html
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x gobl3d module.bash
   ```

2. **Go Module Compilation Failed**
   ```bash
   # Check Go installation
   go version
   
   # Manual compilation
   go build -o gobl3d module.go
   ```

3. **Wordlist Not Found**
   ```bash
   # Install SecLists
   git clone https://github.com/danielmiessler/SecLists.git /opt/SecLists
   
   # Or use system wordlists
   locate wordlist
   ```

4. **Gobuster Not Found**
   ```bash
   # Install gobuster
   sudo apt-get install gobuster
   # or
   go install github.com/OJ/gobuster/v3@latest
   ```

## Tips for Bug Bounty

1. **Start with Quick Recon** to get initial foothold
2. **Use Comprehensive Scan** for deeper enumeration
3. **Integrate with Burp Suite** for manual testing
4. **Chain with other tools** using integration helpers
5. **Monitor output directories** for interesting findings
6. **Use stealth mode** for sensitive targets
7. **Customize wordlists** for specific technologies

## Legal Disclaimer

This tool is for authorized security testing only. Always ensure you have explicit permission before testing any target. The authors are not responsible for any misuse of this tool.

## Credits

- Built on top of gobuster by OJ Reeves
- Inspired by various bug bounty methodologies
- Designed for efficiency and stealth in security testing
