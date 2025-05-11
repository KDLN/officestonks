#!/bin/bash

# Script to diagnose database connectivity issues in Railway

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting MySQL Connection Diagnostics${NC}"
echo -e "${BLUE}===============================${NC}"

# Check environment
echo -e "${YELLOW}Environment Information:${NC}"
echo "Hostname: $(hostname)"
echo "Current user: $(whoami)"
echo "Date: $(date)"
echo "Working directory: $(pwd)"
echo "IP addresses: $(hostname -I 2>/dev/null || echo 'Cannot determine IP')"

# Check if mysql client exists
if command -v mysql &> /dev/null; then
    echo -e "${GREEN}MySQL client found: $(which mysql)${NC}"
    echo "MySQL client version: $(mysql --version)"
else
    echo -e "${RED}MySQL client not found. Installing...${NC}"
    apk add --no-cache mysql-client &> /dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}MySQL client installed successfully: $(which mysql)${NC}"
        echo "MySQL client version: $(mysql --version)"
    else
        echo -e "${RED}Failed to install MySQL client. Some tests will be skipped.${NC}"
    fi
fi

# Check network connectivity
echo -e "${YELLOW}Network Connectivity Tests:${NC}"

# Define database locations to test
hosts=("turntable.proxy.rlwy.net" "caboose.proxy.rlwy.net" "mysql.railway.internal")
ports=("28889" "40558" "3306")

# Test each host:port combination
for host in "${hosts[@]}"; do
    for port in "${ports[@]}"; do
        echo -n "Testing connection to ${host}:${port}... "
        
        # Try with netcat if available
        if command -v nc &> /dev/null; then
            if nc -z -w 5 ${host} ${port} 2>/dev/null; then
                echo -e "${GREEN}SUCCESS${NC}"
                
                # If mysql client exists, try to get banner
                if command -v mysql &> /dev/null; then
                    echo -n "  Getting MySQL banner... "
                    banner=$(echo "exit" | mysql --connect-timeout=5 -h ${host} -P ${port} 2>&1 | grep -i mysql)
                    if [ -n "$banner" ]; then
                        echo -e "${GREEN}${banner}${NC}"
                    else
                        echo -e "${RED}No MySQL banner received${NC}"
                    fi
                fi
            else
                echo -e "${RED}FAILED${NC}"
            fi
        else
            # Try simple timeout command as fallback
            if timeout 5 bash -c "</dev/tcp/${host}/${port}" 2>/dev/null; then
                echo -e "${GREEN}SUCCESS${NC}"
            else
                echo -e "${RED}FAILED${NC}"
            fi
        fi
    done
done

# DNS resolution check
echo -e "${YELLOW}DNS Resolution Tests:${NC}"
for host in "${hosts[@]}"; do
    echo -n "Resolving ${host}... "
    if command -v dig &> /dev/null; then
        ip=$(dig +short ${host} 2>/dev/null)
        if [ -n "$ip" ]; then
            echo -e "${GREEN}${ip}${NC}"
        else
            echo -e "${RED}Failed to resolve${NC}"
        fi
    elif command -v nslookup &> /dev/null; then
        ip=$(nslookup ${host} 2>/dev/null | grep -i address | tail -n1 | awk '{print $2}')
        if [ -n "$ip" ]; then
            echo -e "${GREEN}${ip}${NC}"
        else
            echo -e "${RED}Failed to resolve${NC}"
        fi
    elif command -v host &> /dev/null; then
        ip=$(host ${host} 2>/dev/null | grep -i address | awk '{print $4}')
        if [ -n "$ip" ]; then
            echo -e "${GREEN}${ip}${NC}"
        else
            echo -e "${RED}Failed to resolve${NC}"
        fi
    else
        echo -e "${RED}No DNS tools available${NC}"
    fi
done

# Try various authentication methods if MySQL client exists
if command -v mysql &> /dev/null; then
    echo -e "${YELLOW}MySQL Authentication Tests:${NC}"
    
    # Define credentials to test
    declare -a user_array=("root")
    declare -a password_array=("EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY" "DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE")
    
    # Test environment variables if available
    if [ -n "$MYSQLPASSWORD" ]; then
        password_array+=("$MYSQLPASSWORD")
        echo "Added password from MYSQLPASSWORD environment variable"
    fi
    
    if [ -n "$DB_PASSWORD" ]; then
        password_array+=("$DB_PASSWORD")
        echo "Added password from DB_PASSWORD environment variable"
    fi
    
    # Try each host:port with each credential
    for host in "${hosts[@]}"; do
        for port in "${ports[@]}"; do
            for user in "${user_array[@]}"; do
                for password in "${password_array[@]}"; do
                    echo -n "Testing ${user}@${host}:${port} with password... "
                    if echo "SELECT VERSION();" | mysql --connect-timeout=5 -h ${host} -P ${port} -u ${user} -p${password} 2>/dev/null | grep -q "."; then
                        echo -e "${GREEN}SUCCESS${NC}"
                        echo "  This combination works! Saving to working-credentials.txt"
                        echo "Host: ${host}" > working-credentials.txt
                        echo "Port: ${port}" >> working-credentials.txt
                        echo "User: ${user}" >> working-credentials.txt
                        echo "Password: ${password}" >> working-credentials.txt
                        echo "DSN: ${user}:${password}@tcp(${host}:${port})/railway" >> working-credentials.txt
                        
                        # Try to list tables
                        echo -n "  Listing tables... "
                        tables=$(echo "SHOW TABLES;" | mysql --connect-timeout=5 -h ${host} -P ${port} -u ${user} -p${password} railway 2>/dev/null)
                        if [ $? -eq 0 ]; then
                            echo -e "${GREEN}SUCCESS${NC}"
                            echo "  Tables found:"
                            echo "$tables" | sed 's/^/    /'
                        else
                            echo -e "${RED}FAILED${NC}"
                        fi
                    else
                        echo -e "${RED}FAILED${NC}"
                    fi
                done
            done
        done
    done
else
    echo -e "${RED}MySQL client not available. Skipping authentication tests.${NC}"
fi

# Check environment variables
echo -e "${YELLOW}Environment Variables:${NC}"
echo "MYSQLHOST=${MYSQLHOST:-Not Set}"
echo "MYSQLPORT=${MYSQLPORT:-Not Set}"
echo "MYSQLUSER=${MYSQLUSER:-Not Set}"
echo "MYSQLDATABASE=${MYSQLDATABASE:-Not Set}"
echo "MYSQLPASSWORD=${MYSQLPASSWORD:+Set (value hidden)}"
echo "DB_HOST=${DB_HOST:-Not Set}"
echo "DB_PORT=${DB_PORT:-Not Set}"
echo "DB_USER=${DB_USER:-Not Set}"
echo "DB_NAME=${DB_NAME:-Not Set}"
echo "DB_PASSWORD=${DB_PASSWORD:+Set (value hidden)}"

echo -e "${YELLOW}Diagnostics completed${NC}"
echo -e "${BLUE}===============================${NC}"