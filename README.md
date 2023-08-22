# Command Runner

## Overview

`Command Runner` is a versatile automation tool tailored for executing specific commands within AWS, GCP, and forthcoming support for Azure, as well as on-premises infrastructures. Its primary purpose is to extract, process, and save outputs to designated file paths, making it an invaluable asset for simplifying and enhancing the collection and monitoring of data across various Perforce services.

## Features

- **P4 Instance File Parsing**: Skillfully parses specific Perforce instance files, sanitizing the output to ensure security.
  
- **AWS Token Retrieval**: Safely fetches AWS tokens to facilitate further AWS operations authentication and authorization.
  
- **AWS Instance Identity Data Collection**: Acquires the AWS instance's identity document and associated metadata.
  
- **GCP Instance Identity Data Collection**: Obtains the GCP instance's identity document and related metadata.
  
- **Error Handling**: Effectively logs and conserves any execution errors in a dedicated JSON file.
  
- **Auto-updating from GitHub**: (In the pipeline) The tool will soon be able to autonomously check for, and assimilate, the latest updates from its GitHub repository.

## Prerequisites

- **Golang**: This project fundamentally depends on Golang. Ensure it's both installed and set up correctly.

## Setup & Usage

**Easy Setup/Install**
``curl -sLO https://raw.githubusercontent.com/willKman718/command-runner/master/install.sh && chmod +x install.sh && sudo ./install.sh``

**Manual Setup/Install**
1. **Clone the Repository**:
   ```bash
   git clone https://github.com/willKman718/command-runner.git
   cd /path/to/command-runner/command-runner


2. **Build Binary**:
make
3. Build config
`Default config if not provided /p4/common/config/.push_metrics.cfg`

sample .push_metrics.cfg
```
metrics_host=http://some.ip.or.host:9091
metrics_customer=Customer-Name
metrics_instance=server-name
metrics_user=username-for-pushgateway
metrics_passwd=password-for-pushgateway
report_instance_logfile=/log/file/location   #
```

4. **Setup cmd_config.yaml**
cd configs
vi cmd_config.yaml

sample configs/cmd_config.yaml
```
files:
  - pathtofile: "/etc/hosts"    # Full Path to the File
    monitor_tag: "etc hosts"    # Tag Name for Monitoring
    keywords:                   # List of Keywords to Look for. If parseAll is set to true this is disregarded
      - keyword1                # Line to parse that contains this keyword
      - keyword2                # Line to parse that contains this keyword as well
    parseAll: true              # Set this to true and parse the entire file. If false will only parse lines matching keywords
    parsingLevel: server        # Setting to "server" will parse the entire. "instance" will only run if p4d is installed
    sanitizationKeywords:       # While parsing remove any of these lines containing these words - allows for removal of sensitive info
      - C1A
  - pathtofile: /p4/common/config/p4_%INSTANCE%.vars    # Where %INSTANCE% will replace the p4d instance id.
    monitor_tag: "p4 vars"      # Tag Name for Monitoring
    keywords:
      - export MAILTO
      - export P4USER
      - export P4MASTER_ID
    parseAll: false
    parsingLevel: instance
    sanitizationKeywords:
      - sensitive
      - secret

instance_commands:
  - description: "p4 configure show allservers" # Brief description commands or information about command being ran
    command: "p4 configure show allservers"     # Command to run and gather output from (Beware of escape characters)
    monitor_tag: "p4 configure"                 # Tag Name for Monitoring
  - description: "p4 triggers"
    command: "p4 triggers -o | awk '/^Triggers:/ {flag=1; next} /^$/ {flag=0} flag' | sed 's/^[ \\t]*//'"
    monitor_tag: "p4 triggers"
  - description: "p4 extensions and configs"
    command: "p4 extension --list --type extensions; p4 extension --list --type configs"
    monitor_tag: "p4 extensions"
  - description: "p4 servers"
    command: "p4 servers -J"
    monitor_tag: "p4 servers"
  - description: "p4 property -Al"
    command: "p4 property -Al"
    monitor_tag: "p4 property"
  - description: "p4 -Ztag Without the datefield"
    command: "p4 -Ztag info | awk '!/^... (serverDate|serverUptime)/'"
    monitor_tag: "p4 ztag"

server_commands:
  - description: Server host information        # Brief description commands or information about command being ran
    command: hostnamectl                        # Command to run and gather output from (Beware of escape characters)
    monitor_tag:  hostnamectl                   # Tag Name for Monitoring
  - description: Current active services
    command: systemctl --type=service --state=active
    monitor_tag: systemd
  - description: Current crontab -l
    command: crontab -l
    monitor_tag: crontab
  - description: Swarm Present on this server
    command: "[ -f /opt/perforce/swarm/Version ] && cat /opt/perforce/swarm/Version || true"
    monitor_tag: swarm here
  - description: Swarm URL
    command: "p4 property -n P4.Swarm.URL -l 2>&1 | grep -v 'P4.Swarm.URL - no such property.' || true"
    monitor_tag: swarm url
  - description: SDP Version
    command: "[ -f /p4/sdp/Version ] && cat /p4/sdp/Version || true"
    monitor_tag: SDP Version
```

4. **Create cron Job**
crontab -e
```
10 0 * * * /path/to/command-runner/report_instance_data.sh -c /path/to/.push_metrics.cfg > /dev/null 2>&1 ||:
```

5. **Error Logs**:
Errors are primarily saved in `/p4/1/logs/report_instance_data.log`. Delve into these logs to troubleshoot potential issues or to gain a deeper insight into the processes.




## Future Enhancements

- **Auto-updating**: As mentioned, a future enhancement will be the capability for the tool to check for updates on GitHub and apply them automatically.
  
- **report_instance_data.sh**: Plan to deprecate and remove the `report_instance_data.sh` script in future versions.

- **Fancy p4 commands**: Eliminate the need for long scripts that require escape characters, simplifying command execution.

