Good — this README is already **80% solid**.
What I’m going to do now is **tighten it**, remove anything that sounds self-congratulatory or AI-polished, add the **missing build step**, and make the tone **calm, human, and reviewer-friendly**.

I am **not changing the intent**, only the **clarity and credibility**.

Below is a **clean, submission-ready README** you can paste directly.

---

# LinkedIn Automation POC (Go + Rod)

## Overview

This repository contains a **technical proof-of-concept** for LinkedIn browser automation, implemented in **Go** using the **Rod** browser automation library.

The project focuses on **architecture and execution design**, not growth hacking.
It demonstrates how a browser automation system can be built with:

* Modular Go architecture
* Session persistence
* Human-like interaction patterns
* Anti-detection awareness
* A safe **DRY-RUN execution mode** for evaluation and review

⚠️ **This project is strictly for educational and technical evaluation purposes.**

---

## ⚠️ Important Disclaimer

### Educational Use Only

This project exists to demonstrate:

* Browser automation techniques
* Anti-detection considerations
* Clean system design in Go
* Logging and state management

It is **not intended for real-world automation**.

### Terms of Service Notice

Automating LinkedIn violates LinkedIn’s Terms of Service.
Running automation against real accounts may result in **temporary restrictions or permanent bans**.

**Do not use this project on real LinkedIn accounts.**

---

## Tech Stack

* **Language:** Go
* **Browser Automation:** Rod (Chromium-based)
* **Database:** SQLite (pure Go, no CGO)
* **Logging:** Zap (structured logging)
* **Configuration:** YAML + environment variables
* **Architecture:** Modular `internal/` packages

---

## Project Structure

```
linkedin-automation-poc/
├── cmd/
│   └── app/
│       └── main.go          # Application entry point
│
├── internal/
│   ├── auth/                # Login, session handling, checkpoint detection
│   ├── browser/             # Browser launch & context management
│   ├── config/              # Configuration loading & validation
│   ├── interaction/         # Profile visit & dry-run interaction logic
│   ├── messaging/           # Messaging workflow (dry-run only)
│   ├── search/              # Search execution & result parsing
│   ├── state/               # SQLite persistence (sessions, message state)
│   ├── stealth/             # Human-like behavior & fingerprinting
│   └── log/                 # Logger initialization
│
├── config.yaml              # Application configuration
├── .env.example             # Environment variable template
├── data/
│   └── automation.db        # SQLite database (runtime)
└── README.md
```

---

## Core Capabilities Demonstrated

### 1. Authentication & Session Handling

* Login using credentials from environment variables
* Session cookies persisted in SQLite
* Automatic session reuse across runs
* Detection of LinkedIn security checkpoints (CAPTCHA / verification)

---

### 2. Search & Targeting

* People search using keywords and location
* Page scrolling to load results
* Profile URL extraction from search results
* Duplicate handling via persisted state

---

### 3. Profile Interaction (DRY-RUN)

* Human-like scrolling behavior
* Natural mouse movement
* Hover intent simulation on action buttons
* **No real interactions are executed**

This demonstrates intent and flow without violating platform rules.

---

### 4. Messaging System (DRY-RUN)

* Template-based message structure
* Dynamic variable substitution (name when available, safe fallback otherwise)
* Typing simulation (delays, typos, corrections)
* Message state persistence
* Explicit acknowledgment that real messaging depends on connection acceptance

> Messaging is demonstrated **only in DRY-RUN mode**.
> No messages are actually sent.

---

### 5. Anti-Bot / Stealth Awareness

The project demonstrates awareness of common automation detection signals, including:

* Browser fingerprint randomization
* User-agent and viewport variation
* Human-like mouse movement
* Randomized timing and delays
* Natural scrolling patterns
* Typing rhythm simulation
* Session continuity across runs

These techniques are shown for **educational purposes**, not evasion.

---

## DRY-RUN Mode (Default)

The application runs in **DRY-RUN mode by default**.

In this mode:

* No connection requests are sent
* No messages are delivered
* All actions are logged as **simulated**
* Execution stops intentionally after demonstration

This allows reviewers to evaluate:

* Control flow
* Logging clarity
* Stealth behavior
* Code structure

without performing real automation.

---

## Logging Philosophy

Logs are written as an **execution trace**, not marketing output.

They explicitly show:

* What actions were attempted
* What assumptions exist (e.g., unknown connection state)
* What was intentionally skipped
* Why no real actions occurred

This mirrors how production automation systems are debugged and reviewed.

---

## Setup Instructions

### Prerequisites

* Go 1.21 or newer
* Chromium or Google Chrome installed
* Internet connection

---

### Environment Variables

Create a `.env` file using `.env.example`:

```
LINKEDIN_EMAIL=your_email_here
LINKEDIN_PASSWORD=your_password_here
```

---

### Build the Application

Recommended build command:

```
go build ./cmd/app
```

This produces a single application binary.

---

### Run the Application

```
go run cmd/app/main.go
```

The browser runs in **non-headless mode** so behavior can be observed directly.

---

## Known Limitations (Intentional)

* Messaging is demonstrated without real delivery
* Connection acceptance detection is not implemented
* Profile name extraction depends on LinkedIn DOM and may be unavailable for some results
* CAPTCHA solving is not automated

These constraints are **intentional** to keep the project safe and evaluation-focused.

---

## Purpose of This Repository

This is **not** a scraping tool and **not** a growth automation product.

It exists to demonstrate:

* Browser automation architecture
* Anti-detection awareness
* Clean Go code organization
* Responsible handling of platform constraints

---

## Author

**Thamimul Ansari**
Final-year B.Tech (IT) student

Interests:

* Backend systems
* Browser automation
* Secure and maintainable system design

📧 Email: *thamimul2004@gmail.com*

