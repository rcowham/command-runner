## Comprehensive Guide to the command-runner Binary

### Introduction

The command-runner binary provides a comprehensive toolset for executing a diverse range of commands tailored for system and server monitoring. Its core functionality is to collate data from various sources, structure it in a JSON format, and relay it to the datapushgateway. Typically, it's incorporated with or deployed alongside p4prometheus.

## Features

- **P4 Instance File Parsing**: Skillfully parses specific Perforce instance files, sanitizing the output to ensure security.
  
- **AWS Token Retrieval**: Safely fetches AWS tokens to facilitate further AWS operations authentication and authorization.
  
- **AWS Instance Identity Data Collection**: Acquires the AWS instance's identity document and associated metadata.
  
- **GCP Instance Identity Data Collection**: Obtains the GCP instance's identity document and related metadata.
  
- **Error Handling**: Effectively logs and conserves any execution errors in a dedicated JSON file.
  
### 1. Prerequisites

- Ensure the command-runner binary is installed.
- Validate that you have permission to execute commands on your target servers or SDP/P4 instances.
- Familiarize yourself with the main flags, as they determine the behavior of the binary.

### 2. Understanding the Flags

The command-runner supports numerous flags. Understanding each is vital:

#### Main Flags

- --debug (-d): Enables debug logging, beneficial for troubleshooting.
- --cloud (-c): Indicates the cloud provider, e.g., aws, gcp, azure, or onprem. onprem is the default.
- --instance (-i): Used for SDP instance commands. This flag requires an accompanying instance argument.
- --server (-s): Engages OS-related commands.
- --log (-l): Determines the path to store log files. If unspecified, a default path from the schema will be utilized.
- --output (-o): Denotes the path to the JSON output file. By default, this file is /tmp/out.json.

#### Auxiliary Flags

- --allSDP: Lets the program operate across all SDP instances. This flag is hidden due to safety considerations.
- --autocloud: Activates automatic cloud provider detection. This is concealed for safety purposes.e
- --autobots: Enables running of the autobots scripts. Hidden for safety reasons.
- --mcfg (-m): Gives the path to the metrics configuration file.
- --vars: Denotes the path to the metrics configuration file.
- --cmdcfg (-y): Specifies the location of the cmd\_config.yaml file.
- --nodel: By default, the JSON data output is deleted after execution. This flag prevents that.

### 3. Executing the command-runner

#### Basic Execution

```./command-runner --debug --cloud=aws --instance=Instance123 --server```

This command:

- Activates debug mode for detailed logs.
- Sets the cloud provider as aws.
- Uses Instance123 for SDP instance commands.
- Engages OS-related commands.

#### Advanced Execution

For those looking to use more features, such as automatic cloud detection and ensuring output data isn't deleted:

```./command-runner --debug --autocloud --instance=Instance123 --server --nodel```

### 4. Data Flow & Outputs

- Once executed, the binary assesses flags, preparing the system for data collection.
- Data from various commands, depending on the flags, is gathered and structured into a JSON format.
- The resultant data is then dispatched to datapushgateway.
- You'll find the output JSON in /tmp/out.json, unless an alternative path is specified.
- Logs provide insights into the execution process, assisting in troubleshooting.

### 5. Safety & Best Practices

- Always ensure the right permissions before running the binary, especially in production environments.
- Use hidden flags with caution. They are hidden for safety reasons.
- Regularly check logs, especially when using the --debug flag, to keep tabs on operations.

### 6. Conclusion

The command-runner binary is a potent tool for system and server monitoring. With its array of flags, it offers flexible data collection capabilities. By understanding and leveraging these flags, one can effectively monitor various aspects of their infrastructure and promptly dispatch it to the datapushgateway.