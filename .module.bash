#!/usr/bin/env bash

# Ultimate Gobuster Bug Bounty Tool - Bash Version
# Description: Comprehensive enumeration tool using Gobuster

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_THREADS=30
DEFAULT_DELAY="100ms"
DEFAULT_TIMEOUT="10s"
DEFAULT_EXTENSIONS="php,html,txt,js,asp,aspx,jsp,zip,bak"
DEFAULT_WORDLISTS=(
    "/usr/share/wordlists/dirb/common.txt"
    "/usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt"
    "/usr/share/wordlists/SecLists/Discovery/Web-Content/raft-large-directories.txt"
    "/usr/share/wordlists/SecLists/Discovery/DNS/subdomains-top1million-5000.txt"
)

# Global variables
TARGET=""
OUTPUT_DIR=""
THREADS=$DEFAULT_THREADS
DELAY=$DEFAULT_DELAY
TIMEOUT=$DEFAULT_TIMEOUT
EXTENSIONS=$DEFAULT_EXTENSIONS
WORDLISTS=("${DEFAULT_WORDLISTS[@]}")
SKIP_SSL=true
FOLLOW_REDIRECTS=true
USER_AGENT="Mozilla/5.0 (compatible; UltimateGobuster/1.0)"
PROXY=""
COOKIES=""
HEADERS=()
STATUS_CODES="200,301,302,401,403"
EXCLUDE_CODES="404,500"
RECURSIVE=true
SUBDOMAIN_SCAN=true
VHOST_SCAN=true
DIRECTORY_SCAN=true

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Ultimate Gobuster Bug Bounty Tool - Bash Version

Usage: $0 -t <target> [options]

Required:
  -t, --target <URL>          Target URL (e.g., https://example.com)

Options:
  -o, --output <dir>          Output directory (default: ./gobuster_results_\$(date))
  -T, --threads <num>         Number of threads (default: $DEFAULT_THREADS)
  -d, --delay <time>          Delay between requests (default: $DEFAULT_DELAY)
  -x, --extensions <ext>      File extensions (default: $DEFAULT_EXTENSIONS)
  -w, --wordlist <file>       Custom wordlist (can be used multiple times)
  -p, --proxy <proxy>         Proxy URL (e.g., http://127.0.0.1:8080)
  -c, --cookies <cookies>     Cookies to send
  -H, --header <header>       Custom header (can be used multiple times)
  -a, --user-agent <ua>       User agent string
  -s, --status-codes <codes>  Status codes to match (default: $STATUS_CODES)
  -b, --exclude-codes <codes> Status codes to exclude (default: $EXCLUDE_CODES)
  --timeout <time>            Request timeout (default: $DEFAULT_TIMEOUT)
  --no-ssl-verify             Skip SSL certificate verification
  --no-follow-redirects       Don't follow redirects
  --no-recursive              Skip recursive enumeration
  --no-subdomains             Skip subdomain enumeration
  --no-vhosts                 Skip virtual host enumeration
  --dirs-only                 Only perform directory enumeration
  -h, --help                  Show this help message

Examples:
  $0 -t https://example.com
  $0 -t https://example.com -T 50 -x php,asp,aspx
  $0 -t http://example.com -w /path/to/wordlist.txt --no-recursive
  $0 -t https://example.com -p http://127.0.0.1:8080 -c "session=abc123"

EOF
}

# Function to parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--target)
                TARGET="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -T|--threads)
                THREADS="$2"
                shift 2
                ;;
            -d|--delay)
                DELAY="$2"
                shift 2
                ;;
            -x|--extensions)
                EXTENSIONS="$2"
                shift 2
                ;;
            -w|--wordlist)
                WORDLISTS=("$2")
                shift 2
                ;;
            -p|--proxy)
                PROXY="$2"
                shift 2
                ;;
            -c|--cookies)
                COOKIES="$2"
                shift 2
                ;;
            -H|--header)
                HEADERS+=("$2")
                shift 2
                ;;
            -a|--user-agent)
                USER_AGENT="$2"
                shift 2
                ;;
            -s|--status-codes)
                STATUS_CODES="$2"
                shift 2
                ;;
            -b|--exclude-codes)
                EXCLUDE_CODES="$2"
                shift 2
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --no-ssl-verify)
                SKIP_SSL=true
                shift
                ;;
            --no-follow-redirects)
                FOLLOW_REDIRECTS=false
                shift
                ;;
            --no-recursive)
                RECURSIVE=false
                shift
                ;;
            --no-subdomains)
                SUBDOMAIN_SCAN=false
                shift
                ;;
            --no-vhosts)
                VHOST_SCAN=false
                shift
                ;;
            --dirs-only)
                SUBDOMAIN_SCAN=false
                VHOST_SCAN=false
                RECURSIVE=false
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Function to validate dependencies
check_dependencies() {
    if ! command -v gobuster &> /dev/null; then
        print_error "gobuster not found. Please install gobuster first."
        exit 1
    fi
    
    print_success "gobuster found: $(gobuster version 2>/dev/null | head -n1 || echo 'Version unknown')"
}

# Function to validate wordlists
validate_wordlists() {
    local valid_wordlists=()
    
    for wordlist in "${WORDLISTS[@]}"; do
        if [[ -f "$wordlist" ]]; then
            valid_wordlists+=("$wordlist")
            print_success "Found wordlist: $wordlist"
        else
            print_warning "Wordlist not found: $wordlist"
        fi
    done
    
    if [[ ${#valid_wordlists[@]} -eq 0 ]]; then
        print_error "No valid wordlists found!"
        exit 1
    fi
    
    WORDLISTS=("${valid_wordlists[@]}")
}

# Function to build common gobuster arguments
build_common_args() {
    local args=()
    
    [[ -n "$THREADS" ]] && args+=("-t" "$THREADS")
    [[ -n "$DELAY" ]] && args+=("--delay" "$DELAY")
    [[ -n "$TIMEOUT" ]] && args+=("--timeout" "$TIMEOUT")
    [[ "$SKIP_SSL" == true ]] && args+=("-k")
    [[ "$FOLLOW_REDIRECTS" == true ]] && args+=("-r")
    [[ -n "$USER_AGENT" ]] && args+=("-a" "$USER_AGENT")
    [[ -n "$PROXY" ]] && args+=("--proxy" "$PROXY")
    [[ -n "$COOKIES" ]] && args+=("-c" "$COOKIES")
    [[ -n "$STATUS_CODES" ]] && args+=("-s" "$STATUS_CODES")
    [[ -n "$EXCLUDE_CODES" ]] && args+=("-b" "$EXCLUDE_CODES")
    
    for header in "${HEADERS[@]}"; do
        args+=("-H" "$header")
    done
    
    echo "${args[@]}"
}

# Function to extract domain from URL
extract_domain() {
    local url="$1"
    # Remove protocol
    url="${url#http://}"
    url="${url#https://}"
    # Remove path and port
    url="${url%%/*}"
    url="${url%%:*}"
    echo "$url"
}

# Function to run directory enumeration
run_directory_enumeration() {
    if [[ "$DIRECTORY_SCAN" != true ]]; then
        return
    fi
    
    print_info "Starting directory enumeration..."
    
    local counter=1
    for wordlist in "${WORDLISTS[@]}"; do
        if [[ ! -f "$wordlist" ]]; then
            continue
        fi
        
        local output_file="$OUTPUT_DIR/directories_${counter}.txt"
        local args=("dir" "-u" "$TARGET" "-w" "$wordlist" "-o" "$output_file")
        
        # Add common arguments
        read -ra common_args <<< "$(build_common_args)"
        args+=("${common_args[@]}")
        
        # Add extensions
        [[ -n
