(() => {
    "use strict";

    const State = Object.freeze({
        IDLE: "idle",
        CONNECTING: "connecting",
        CONNECTED: "connected",
        ERROR: "error"
    });

    const terminalColorVariables = {
        background: "--color-base-100",
        foreground: "--color-base-content",
        cursor: "--color-base-content",
        selection: "--color-primary-content",
        black: "--color-neutral",
        red: "--color-error",
        green: "--color-success",
        yellow: "--color-warning",
        blue: "--color-primary",
        magenta: "--color-secondary",
        cyan: "--color-accent",
        white: "--color-base-100",
        brightBlack: "--color-neutral-content",
        brightRed: "--color-error",
        brightGreen: "--color-success",
        brightYellow: "--color-warning",
        brightBlue: "--color-primary",
        brightMagenta: "--color-secondary",
        brightCyan: "--color-accent",
        brightWhite: "--color-base-content"
    };

    const elements = {
        connectButton: document.getElementById("connectBtn"),
        connectionForm: document.getElementById("connectionForm"),
        closeModalButton: document.getElementById("closeModalBtn"),
        host: document.getElementById("host"),
        port: document.getElementById("port"),
        username: document.getElementById("username"),
        password: document.getElementById("password"),
        privateKey: document.getElementById("privateKey"),
        passphrase: document.getElementById("passphrase"),
        modal: document.getElementById("connectionModal"),
        statusIndicator: document.getElementById("statusIndicator"),
        statusText: document.getElementById("statusText"),
        terminal: document.getElementById("terminal"),
        terminalContainer: document.querySelector(".terminal-container"),
        authTabs: [...document.querySelectorAll(".auth-tab")],
        authPanels: [...document.querySelectorAll(".auth-content")]
    };

    const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {}
    });
    const fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(elements.terminal);

    let state = State.IDLE;
    let socket = null;
    let authMethod = "password";

    function cssColorToHex(value, fallback) {
        const parsed = window.culori?.parse?.(value);
        return parsed ? window.culori.formatHex(parsed) : fallback;
    }

    function terminalThemeFromCSS() {
        const styles = getComputedStyle(document.body);
        return Object.fromEntries(Object.entries(terminalColorVariables).map(([name, variable]) => [
            name,
            cssColorToHex(styles.getPropertyValue(variable).trim(), "#000000")
        ]));
    }

    function applyTheme(theme) {
        document.body.classList.toggle("dark", theme === "dark");
        document.body.classList.toggle("light", theme !== "dark");
        term.options.theme = terminalThemeFromCSS();
    }

    function setState(nextState, message) {
        state = nextState;
        elements.statusText.textContent = message;
        elements.statusIndicator.className = "status-indicator";

        const isConnected = nextState === State.CONNECTED;
        const isError = nextState === State.ERROR;
        const isConnecting = nextState === State.CONNECTING;
        elements.statusIndicator.classList.toggle("connected", isConnected);
        elements.statusIndicator.classList.toggle("error", isError);
        elements.connectButton.textContent = isConnected ? "Disconnect" : isConnecting ? "Connecting…" : "Connect";
        elements.connectButton.disabled = isConnecting;
        elements.connectButton.classList.toggle("btn-primary", !isConnected);
        elements.connectButton.classList.toggle("btn-error", isConnected);
    }

    function writeWelcome() {
        term.clear();
        term.write("Terminal - Click Connect to start an SSH session.\\r\\n");
    }

    function activeSocket(candidate) {
        return socket === candidate;
    }

    function closeSocket(candidate = socket) {
        if (!candidate || !activeSocket(candidate)) {
            return;
        }
        socket = null;
        candidate.disconnect();
    }

    function fitTerminal() {
        requestAnimationFrame(() => {
            fitAddon.fit();
            if (state === State.CONNECTED && socket) {
                socket.emit("resize", { cols: term.cols, rows: term.rows });
            }
        });
    }

    function setAuthMethod(nextMethod) {
        authMethod = nextMethod;
        elements.authTabs.forEach(tab => {
            const active = tab.dataset.auth === nextMethod;
            tab.classList.toggle("active", active);
            tab.setAttribute("aria-selected", String(active));
        });
        elements.authPanels.forEach(panel => {
            const active = panel.id === `${nextMethod}-auth`;
            panel.classList.toggle("active", active);
            panel.hidden = !active;
        });
    }

    function requestFromForm() {
        const host = elements.host.value.trim();
        const port = elements.port.value.trim();
        const username = elements.username.value.trim();
        if (!host || !port || !username) {
            throw new Error("Host, port, and username are required");
        }

        const request = { host, port, username };
        if (authMethod === "password") {
            request.password = elements.password.value;
        } else {
            request.privateKey = elements.privateKey.value.trim();
            request.passphrase = elements.passphrase.value;
        }
        return request;
    }

    function handleConnect() {
        if (state === State.CONNECTING || state === State.CONNECTED) {
            return;
        }

        let request;
        try {
            request = requestFromForm();
        } catch (error) {
            setState(State.ERROR, error.message);
            return;
        }

        setState(State.CONNECTING, "Connecting…");
        const nextSocket = io("/ssh", {
            transports: ["websocket", "polling"],
            reconnection: false
        });
        socket = nextSocket;

        nextSocket.on("connect", () => {
            if (activeSocket(nextSocket)) {
                nextSocket.emit("connect_ssh", request);
            }
        });

        nextSocket.on("ssh_connected", data => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            setState(State.CONNECTED, `Connected to ${data.user}@${data.host}:${data.port}`);
            elements.modal.close();
            term.clear();
            term.write(`Connected to ${data.user}@${data.host}:${data.port}\\r\\n`);
            term.focus();
            fitTerminal();
        });

        nextSocket.on("ssh_error", error => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            setState(State.ERROR, `Error: ${error}`);
            closeSocket(nextSocket);
        });

        nextSocket.on("ssh_disconnected", message => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            setState(State.IDLE, message || "Not connected");
            term.write(`\\r\\n${message || "SSH session closed"}\\r\\n`);
            closeSocket(nextSocket);
        });

        nextSocket.on("terminal_output", data => {
            if (activeSocket(nextSocket)) {
                term.write(data);
            }
        });

        nextSocket.on("connect_error", error => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            setState(State.ERROR, `Connection failed: ${error.message}`);
            closeSocket(nextSocket);
        });

        nextSocket.on("disconnect", reason => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            socket = null;
            if (state === State.CONNECTED || state === State.CONNECTING) {
                setState(State.IDLE, `Disconnected${reason ? `: ${reason}` : ""}`);
                term.write("\\r\\nSSH session closed\\r\\n");
            }
        });
    }

    function disconnect() {
        const wasConnected = state === State.CONNECTED;
        setState(State.IDLE, "Not connected");
        closeSocket();
        if (wasConnected) {
            writeWelcome();
        }
    }

    function toggleConnection() {
        if (state === State.CONNECTED || state === State.CONNECTING) {
            disconnect();
            return;
        }
        if (!elements.modal.open) {
            elements.modal.showModal();
            elements.host.focus();
        }
    }

    elements.connectButton.addEventListener("click", toggleConnection);
    elements.closeModalButton.addEventListener("click", () => elements.modal.close());
    elements.connectionForm.addEventListener("submit", event => {
        event.preventDefault();
        handleConnect();
    });
    elements.authTabs.forEach(tab => tab.addEventListener("click", () => setAuthMethod(tab.dataset.auth)));
    elements.modal.addEventListener("click", event => {
        if (event.target === elements.modal) {
            elements.modal.close();
        }
    });
    term.onData(data => {
        if (state === State.CONNECTED && socket) {
            socket.emit("terminal_input", data);
        }
    });

    const resizeObserver = new ResizeObserver(fitTerminal);
    resizeObserver.observe(elements.terminalContainer);
    window.addEventListener("beforeunload", () => closeSocket());
    if (window.webSDK) {
        window.webSDK.onThemeChange(applyTheme);
        applyTheme(window.webSDK.getTheme());
    } else {
        applyTheme("light");
    }
    setState(State.IDLE, "Not connected");
    writeWelcome();
    fitTerminal();
    term.focus();
})();
