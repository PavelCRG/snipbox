import { useState } from "react";

const collections = [
  {
    name: "Git",
    count: "12 commands",
    color: "orange",
    detail:
      "Branches, staging, commits, and recovery without hunting through old notes.",
  },
  {
    name: "PowerShell",
    count: "10 commands",
    color: "blue",
    detail:
      "Processes, services, ports, files, and the everyday Windows toolkit.",
  },
  {
    name: "Docker",
    count: "6 commands",
    color: "cyan",
    detail:
      "Inspect containers, follow logs, clean up, and build images with confidence.",
  },
  {
    name: "Linux / Bash",
    count: "6 commands",
    color: "green",
    detail: "Keep useful file, process, and system commands close at hand.",
  },
  {
    name: "Network",
    count: "6 commands",
    color: "purple",
    detail:
      "Reach for practical connection and DNS diagnostics when you need them.",
  },
];

const previewCommands = {
  Git: [
    "Check working tree",
    "Create a branch",
    "Review recent commits",
    "Undo last commit",
  ],
  PowerShell: [
    "Find a listening port",
    "List running processes",
    "Search file contents",
    "Restart a service",
  ],
  Docker: [
    "List running containers",
    "Follow container logs",
    "Build an image",
    "Open a shell",
  ],
  "Linux / Bash": [
    "Find large files",
    "Inspect disk usage",
    "Search command history",
    "Show memory usage",
  ],
  Network: [
    "Check DNS records",
    "Test a TCP port",
    "Show route to host",
    "Inspect local addresses",
  ],
};

const previewDetails = {
  Git: {
    category: "Branches & merging",
    title: "Create a branch",
    description: "Start a new branch and switch to it in one step.",
    before: "git switch -c ",
    value: "feature/new-work",
    variable: "branch",
  },
  PowerShell: {
    category: "Processes & ports",
    title: "List running processes",
    description: "Sort processes by CPU time to spot busy applications.",
    command: "Get-Process | Sort-Object CPU -Descending",
  },
  Docker: {
    category: "Containers & logs",
    title: "Follow container logs",
    description: "Stream new output from a container as it runs.",
    before: "docker logs -f ",
    value: "app",
    variable: "container",
  },
  "Linux / Bash": {
    category: "System & resources",
    title: "Inspect disk usage",
    description: "Show filesystem space in a readable format.",
    command: "df -h",
  },
  Network: {
    category: "Connection checks",
    title: "Test a TCP port",
    description: "Check whether a host accepts a connection on a port.",
    before: "nc -vz ",
    value: "example.com 443",
    variable: "host and port",
  },
};

function Arrow() {
  return <span aria-hidden="true">↗</span>;
}

function TerminalPreview() {
  const [active, setActive] = useState("Git");
  const [copied, setCopied] = useState(false);
  const detail = previewDetails[active];
  const command = detail.command ?? `${detail.before}${detail.value}`;

  async function copyCommand() {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div
      className="terminal"
      aria-label="Illustration of the Snipbox command browser"
    >
      <div className="terminal-bar">
        <div className="window-dots" aria-hidden="true">
          <i />
          <i />
          <i />
        </div>
        <span>
          SNIPBOX <span className="terminal-muted">/ command library</span>
        </span>
        <span className="terminal-platform">LOCAL</span>
      </div>
      <div className="terminal-search">
        <span>⌕</span> Search commands <kbd>type to filter</kbd>
      </div>
      <div
        className="terminal-tabs"
        role="tablist"
        aria-label="Command collections"
      >
        {collections.map(({ name }) => (
          <button
            key={name}
            type="button"
            role="tab"
            aria-selected={active === name}
            className={active === name ? "selected" : ""}
            onClick={() => setActive(name)}
          >
            {name === "Linux / Bash" ? "Linux" : name}
          </button>
        ))}
      </div>
      <div className="terminal-body">
        <div
          className="command-list"
          role="tabpanel"
          aria-label={`${active} commands`}
        >
          {previewCommands[active].map((item, index) => (
            <div
              key={item}
              className={`command-row ${index === 1 ? "current" : ""}`}
            >
              <span className="row-symbol">{index === 1 ? "›" : " "}</span>
              <span>{item}</span>
              {index === 1 && <span className="row-count">3×</span>}
            </div>
          ))}
        </div>
        <div className="command-detail">
          <div className="detail-label">{detail.category}</div>
          <h3>{detail.title}</h3>
          <p>{detail.description}</p>
          <div className="code-block">
            <code>
              {detail.variable ? (
                <>
                  {detail.before}
                  <span>{detail.value}</span>
                </>
              ) : (
                detail.command
              )}
            </code>
          </div>
          <div className="detail-foot">
            <span>
              {detail.variable
                ? `VARIABLE · ${detail.variable}`
                : "EXAMPLE · READY TO USE"}
            </span>
            <button
              type="button"
              onClick={copyCommand}
              aria-label="Copy example command"
            >
              {copied ? "Copied" : "Copy"}{" "}
              <span aria-hidden="true">{copied ? "✓" : "⧉"}</span>
            </button>
          </div>
        </div>
      </div>
      <div className="terminal-foot">
        <span>↑↓ navigate</span>
        <span>↵ copy</span>
        <span>ctrl+r run</span>
        <span>esc close</span>
      </div>
    </div>
  );
}

function App() {
  return (
    <main>
      <header className="site-header page-width">
        <a className="brand" href="#top" aria-label="Snipbox home">
          <span className="brand-mark">s.</span>
          <span>snipbox</span>
        </a>
        <nav aria-label="Main navigation">
          <a href="#collections">Collections</a>
          <a href="#how-it-works">How it works</a>
          <a
            className="nav-github"
            href="https://github.com/PavelCRG/snipbox"
            target="_blank"
            rel="noreferrer"
          >
            GitHub <Arrow />
          </a>
        </nav>
      </header>

      <section className="hero page-width" id="top">
        <div className="hero-copy">
          <p className="eyebrow">
            <span /> A calmer command line
          </p>
          <h1>
            Your commands,
            <br />
            <em>ready when you are.</em>
          </h1>
          <p className="hero-lede">
            A searchable home for the terminal commands you rely on. Find the
            right one, understand what it does, then copy or run it without
            breaking your flow.
          </p>
          <div className="hero-actions">
            <a className="button-primary" href="#getting-started">
              Get started <Arrow />
            </a>
            <a
              className="text-link"
              href="https://github.com/PavelCRG/snipbox"
              target="_blank"
              rel="noreferrer"
            >
              Explore the source <Arrow />
            </a>
          </div>
          <div className="hero-meta">
            <span>Rust search engine</span>
            <i /> <span>Go terminal interface</span>
            <i /> <span>Runs locally</span>
          </div>
        </div>
        <TerminalPreview />
        <div className="hero-index" aria-hidden="true">
          <span>01</span>
          <i /> COMMANDS, WITHOUT THE DETOUR
        </div>
      </section>

      <section className="collections-section" id="collections">
        <div className="page-width">
          <div className="section-heading">
            <div>
              <p className="eyebrow">A useful place to start</p>
              <h2>
                Five collections.
                <br />
                <em>Room for your own.</em>
              </h2>
            </div>
            <p className="section-intro">
              Start with a carefully organized library for the tools you already
              use. Search across every command, or browse the tree by category
              and topic.
            </p>
          </div>
          <div className="collection-grid">
            {collections.map((item, index) => (
              <article
                className={`collection collection-${item.color}`}
                key={item.name}
              >
                <div className="collection-top">
                  <span className="collection-number">0{index + 1}</span>
                  <span className="collection-count">{item.count}</span>
                </div>
                <h3>{item.name}</h3>
                <p>{item.detail}</p>
                <span className="collection-line" />
              </article>
            ))}
          </div>
          <p className="collection-note">
            <span>40</span> useful starters, each with a clear description,
            tags, and editable template values.
          </p>
        </div>
      </section>

      <section className="workflow page-width" id="how-it-works">
        <div className="workflow-title">
          <p className="eyebrow">A focused workflow</p>
          <h2>
            From “what was that command?”
            <br />
            <em>to done.</em>
          </h2>
          <p>
            Snipbox keeps discovery, context, and action together, so useful
            commands are easier to reuse and safer to run.
          </p>
        </div>
        <div className="steps">
          <article>
            <span>01 / FIND</span>
            <h3>Search the way you think.</h3>
            <p>
              Type a fragment from the name, description, tag, or command. Fuzzy
              search brings the closest match into view as you type.
            </p>
          </article>
          <article>
            <span>02 / UNDERSTAND</span>
            <h3>See the details first.</h3>
            <p>
              Read what a command does, inspect its flags, and fill in named
              values like <code>{"{{branch:feature/new-work}}"}</code>.
            </p>
          </article>
          <article>
            <span>03 / USE</span>
            <h3>Copy it or run it.</h3>
            <p>
              Copy a ready-to-use command, or run it from the terminal. Snipbox
              asks you to confirm commands that match destructive patterns.
            </p>
          </article>
        </div>
      </section>

      <section className="principles">
        <div className="page-width principle-grid">
          <div>
            <p className="eyebrow">Designed to stay yours</p>
            <h2>
              Your library stays
              <br />
              <em>on your machine.</em>
            </h2>
          </div>
          <p className="principle-copy">
            Commands and favorites live in a readable TOML file in your home
            directory. History import is opt-in and local. There is no account
            to create and no service sending your snippets elsewhere.
          </p>
          <div className="principle-facts">
            <div>
              <span>01</span>
              <strong>Plain TOML</strong>
              <small>Easy to inspect and back up</small>
            </div>
            <div>
              <span>02</span>
              <strong>Local history</strong>
              <small>Imported only when you ask</small>
            </div>
            <div>
              <span>03</span>
              <strong>Portable sync</strong>
              <small>Use your own Git remote</small>
            </div>
          </div>
        </div>
      </section>

      <section className="start page-width" id="getting-started">
        <div>
          <p className="eyebrow">Open source · by Pavel Slepcov</p>
          <h2>
            Make your next
            <br />
            <em>command count.</em>
          </h2>
          <p>
            Build Snipbox on your machine, launch the terminal workspace, and
            make the starter library your own.
          </p>
          <div className="start-links">
            <a
              className="button-primary"
              href="https://github.com/PavelCRG/snipbox#build-from-source"
              target="_blank"
              rel="noreferrer"
            >
              Build instructions <Arrow />
            </a>
            <a
              className="text-link"
              href="https://github.com/PavelCRG/snipbox"
              target="_blank"
              rel="noreferrer"
            >
              View on GitHub <Arrow />
            </a>
          </div>
        </div>
        <div className="install-card">
          <div className="install-head">
            <span>FROM THE PROJECT ROOT</span>
            <span>WINDOWS · LINUX · macOS</span>
          </div>
          <code>
            <i>$</i> make build
          </code>
          <div className="install-rule" />
          <p>
            Then launch the built binary and search your library as you work.
          </p>
          <code className="launch">
            <i>$</i> ./bin/snipbox
          </code>
          <small>
            On Windows: <code>.\bin\snipbox.exe</code>
          </small>
        </div>
      </section>

      <footer className="site-footer page-width">
        <a className="brand" href="#top">
          <span className="brand-mark">s.</span>
          <span>snipbox</span>
        </a>
        <span>A small home for the commands that keep work moving.</span>
        <a
          href="https://github.com/PavelCRG/snipbox"
          target="_blank"
          rel="noreferrer"
        >
          Source on GitHub <Arrow />
        </a>
      </footer>
    </main>
  );
}

export default App;
