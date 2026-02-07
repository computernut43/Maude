# In SUPER BETA MODE, as I just "Vibe" coded this since I wanted to use MariaDB, I have not used it yet, so I cannot confirm that this works or not. Use at your own risk. 

# Maude

**Listen here, darling — this is a robust, opinionated issue tracker that doesn't beat around the bush.**

**Platforms:** macOS, Linux, Windows, FreeBSD *(because equality means running everywhere)*

[![License](https://img.shields.io/github/license/steveyegge/beads)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/steveyegge/beads)](https://goreportcard.com/report/github.com/steveyegge/beads)
[![Release](https://img.shields.io/github/v/release/steveyegge/beads)](https://github.com/steveyegge/beads/releases)
[![npm version](https://img.shields.io/npm/v/@beads/bd)](https://www.npmjs.com/package/@beads/bd)
[![PyPI](https://img.shields.io/pypi/v/beads-mcp)](https://pypi.org/project/beads-mcp/)

*God'll get you for that, Walter* — and He'll get you for using sticky notes to track your tasks too. Maude provides a persistent, structured memory for coding agents. It replaces those messy markdown plans with a proper dependency-aware graph, because some of us believe in *organization* and *doing things right*.

## ⚡ Quick Start (And I Do Mean Quick)

```bash
# Install the CLI — we don't have all day, dear
curl -fsSL https://raw.githubusercontent.com/computernut43/Maude/main/scripts/install.sh | bash

# Initialize in YOUR project
cd your-project
bd init

# Tell your agent who's in charge
echo "Use 'bd' for task tracking" >> AGENTS.md
```

**Important:** Maude is a CLI tool you install once and use everywhere. You don't need to clone this repository into your project. *Honestly, some people...*

## 🛠 Features (The Good Stuff)

* **MariaDB Backend:** Issues stored in MariaDB for robust, server-based storage with multi-user support. *Because we believe in a solid foundation, not some flimsy file on disk.*
* **Agent-Optimized:** JSON output, dependency tracking, and auto-ready task detection. *It knows what you need before you do.*
* **Zero Conflict:** Hash-based IDs (`bd-a1b2`) prevent merge collisions in multi-agent/multi-branch workflows. *No more fighting — well, at least not about merge conflicts.*
* **Invisible Infrastructure:** Background daemon for auto-sync; optional JSONL export for version control. *It does the work so you don't have to think about it.*
* **Compaction:** Semantic "memory decay" summarizes old closed tasks to save context window. *Out with the old, in with the relevant.*

## 📖 Essential Commands (The Ones That Matter)

| Command | What It Does |
| --- | --- |
| `bd ready` | List tasks with no open blockers. *The ones you can actually work on.* |
| `bd create "Title" -p 0` | Create a P0 task. *For when it's actually important.* |
| `bd update <id> --claim` | Atomically claim a task (sets assignee + in_progress). *Call dibs, officially.* |
| `bd dep add <child> <parent>` | Link tasks (blocks, related, parent-child). *Everything is connected, dear.* |
| `bd show <id>` | View task details and audit trail. *Get the full story.* |

## 🔗 Hierarchy & Workflow (Getting Organized)

Maude supports hierarchical IDs for epics, because *structure matters*:

* `bd-a3f8` (Epic) — *The big picture*
* `bd-a3f8.1` (Task) — *The work that needs doing*
* `bd-a3f8.1.1` (Sub-task) — *The nitty-gritty details*

**Stealth Mode:** Run `bd init --stealth` to use Maude locally without committing files to the main repo. *For when you need to work in peace without the whole world knowing your business.*

**Contributor vs Maintainer:** When working on open-source projects:

* **Contributors** (forked repos): Run `bd init --contributor` to route planning issues to a separate repo (e.g., `~/.beads-planning`). *Keeps your experimental ideas out of the official PRs. You're welcome.*
* **Maintainers** (write access): Maude auto-detects maintainer role via SSH URLs or HTTPS with credentials. Only need `git config beads.role maintainer` if using GitHub HTTPS without credentials but you have write access. *Smart enough to figure it out on its own, most of the time.*

## 📦 Installation (Let's Get This Show on the Road)

* **npm:** `npm install -g @beads/bd`
* **Homebrew:** `brew install beads`
* **Go:** `go install github.com/steveyegge/beads/cmd/bd@latest`

**Requirements:** Linux, FreeBSD, macOS, or Windows. *We don't discriminate.*

## 🌐 Community Tools (The People Have Spoken)

See [docs/COMMUNITY_TOOLS.md](docs/COMMUNITY_TOOLS.md) for a curated list of community-built UIs, extensions, and integrations — including terminal interfaces, web UIs, editor extensions, and native apps. *It takes a village, after all.*

## 📝 Documentation (Read It. Seriously.)

* [Installing](docs/INSTALLING.md) | [MariaDB Setup](docs/MARIADB.md) | [Agent Workflow](AGENT_INSTRUCTIONS.md) | [Copilot Setup](docs/COPILOT_INTEGRATION.md) | [Articles](ARTICLES.md) | [Sync Branch Mode](docs/PROTECTED_BRANCHES.md) | [Troubleshooting](docs/TROUBLESHOOTING.md) | [FAQ](docs/FAQ.md)
* [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/steveyegge/beads)

---

*"I'm not opinionated, I'm just always right."* — Maude
