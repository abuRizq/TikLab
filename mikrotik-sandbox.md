# MikroTik Sandbox CLI (TikLab)

## 1. Product Overview

**Product Name:** MikroTik Sandbox CLI
**Description:** A software tool that generates virtual testing environments simulating MikroTik devices (RouterOS), pre-loaded with users, network configurations, and realistic data traffic.
**Objective:** To provide a safe environment for developers to test billing systems, Hotspot managers, and automation tools under realistic loads without requiring physical hardware or risking production networks.

### 1.1 Target Audience

* SaaS developers building ISP/Hotspot billing systems.
* Network automation and software integration engineers.

### 1.2 Value Proposition & Advantages

* **Realistic Data:** Eliminates the "scaling gap" where scripts work for 10 users but fail or lag when moving to 1,000+ users.
* **Safety & Risk Mitigation:** Allows for the testing of destructive or sensitive commands (e.g., user disconnection or configuration overrides) without affecting real customers.
* **Hardware Independence:** Enables development and testing from any location without the need for physical lab routers.
* **Zero-Friction Deployment:** A true "plug-and-play" experience. Available via Docker Hub or GitHub, the entire environment can be spun up using a **single command**, eliminating complex installation guides and dependencies.

---

## 2. Beta Scope: Operating Modes

The Beta version focuses strictly on **Synthetic Generation** to ensure immediate environment readiness.

### 2.1 Synthetic Mode (Default Mode)

Generates a complete network environment from scratch based on a predefined profile.

* **Function:** Builds a network simulating a small service provider (approx. 50 active users).
* **Enabled Services:** Includes ready-to-use configurations for DHCP and Hotspot services.

---

## 3. Core Functional Features (Beta)

### 3.1 Subscriber Service Simulation

* **Access Management (Hotspot):** Full support for simulating user logins via browser portals or device-based authentication.
* **Address Allocation (DHCP):** Simulates client devices requesting and receiving IP addresses dynamically, mirroring real-world behavior.
* **Bandwidth Control (Queues):** Implements realistic bandwidth limits on users and links them to their data consumption behavior.

### 3.2 User Behavior Profiles

The Beta is designed to reflect real consumption patterns divided as follows:

1. **Idle Users (40%):** Minimal activity (pings/DNS) just to keep the session active.
2. **Standard Browsing (45%):** Typical web surfing and small file downloads.
3. **Heavy Users (15%):** Continuous, high-bandwidth consumption (heavy downloads/streaming).

---

## 4. Functional Architecture

The architecture focuses on isolating the test environment and ensuring the system responds exactly like an independent physical device.

* **Control Unit:** Receives user commands and manages the sandbox lifecycle.
* **RouterOS Simulator:** Provides a full-featured operating system with all its software capabilities.
* **Human Behavior Engine:** Generates data traffic that mimics the activity of hundreds of concurrent users.
* **State Management System:** Ensures the ability to return to a "clean slate" at any time without reinstallation.

### 4.1 Access & Integration

The tool provides standard access points identical to physical hardware:

* **API Access:** For programmatic integration.
* **Graphical Access (Winbox):** For manual configuration and monitoring.
* **Terminal Access (SSH):** For command-line management.

### 4.2 Technology Stack & Distribution

* **Core Language (Go):** The entire CLI and control unit are engineered in **Go (Golang)**. This guarantees high-speed execution, minimal resource consumption, and native cross-platform compatibility as a standalone binary.
* **Single-Command Delivery:** Distributed primarily via Docker Hub and GitHub releases. Users can deploy and run the full environment flawlessly with a single line of code (e.g., a unified `docker run` command), regardless of their underlying host system.

---

## 5. Core Operations (CLI)

All operations are designed to be executed via a single, intuitive binary or Docker command (e.g., using the `tiklab` prefix).

| Command | Functional Purpose |
| --- | --- |
| **`tiklab create`** | Builds the virtual environment and prepares initial settings. |
| **`tiklab start`** | Activates the lab and begins generating user traffic. |
| **`tiklab reset`** | Instantly reverts the router to its original clean state (wiping all test changes). |
| **`tiklab destroy`** | Completely removes the environment to free up system resources. |
| **`tiklab scale`** | Dynamically increases or decreases the number of active simulated users during a test. |

*(Note: Replaced `sandbox` with `tiklab` in the table to better reflect the branding, but this can easily be reverted if you prefer the generic term).*

---

## 6. Future Roadmap (Out of Scope for Beta)

These high-value features are planned for subsequent releases:

* **PPPoE Simulation:** To support testing for home broadband and FTTH networks.
* **Mirror Mode:** Importing configurations from an existing physical router and converting them into a virtual lab.
* **Multiple Traffic Profiles:** Selecting specific network types (e.g., Hotel Wi-Fi, School, or Wireless ISP).
* **Local LAN Integration:** Connecting the sandbox directly to the developer's physical office network.
* **External RADIUS Integration:** Testing links with external authentication servers.
* **Network Attack Simulation:** Testing firewall resilience against breaches and port scans.
* **Monitoring Dashboards:** Visual interfaces to track lab performance and data consumption.

---

## 7. Success Metrics

* **Cross-Platform Consistency:** Successful deployment and identical behavior on Windows, Linux, and macOS.
* **Performance Under Load:** The API remains responsive even during bandwidth saturation.
* **Simulation Accuracy:** User consumption ratios are correctly reflected in management interfaces.
* **Iteration Speed:** The reset command works fast and consistently.
