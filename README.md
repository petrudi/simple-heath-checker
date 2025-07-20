# Go Service Checker

This is a simple Go program to check the availability of services (HTTP(S), TCP, Ping) listed in a configuration file. It prints whether each service is `UP` or `DOWN`, along with the reason.

## Features

- Reads list of services from a YAML config file
- Checks services via:
  - HTTP/HTTPS
  - TCP
  - Ping (fallback)
- Uses Go routines for concurrency

---

## Usage

1. **Clone the repo**  
   ```bash
   git clone https://github.com/petrudi/simple-heath-checker 
   cd simple-heath-checker
   ```
2. **Create your config file**

Copy the provided sample and modify it.

Edit `config.yaml` and add your services.


3. **Run the health checker**
    ```bash
    go run main.go
    ```

4. **How to install on linux**
    ```bash
    mkdir -p ~/.health-checker
    cp config.sample.yaml ~/.health-checker/config.yaml
    ```
    run it using the following command:
    ```bash
    health-checker
    #or
    health-checker -c /path/to/config.yaml
    ```



**Note**
Your real config.yaml is ignored via .gitignore for safety.

Only config.sample.yaml is committed to the repo as a template.


