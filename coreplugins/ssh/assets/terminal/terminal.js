(() => {
    "use strict";

    const State = Object.freeze({
        IDLE: "idle",
        CONNECTING: "connecting",
        CONNECTED: "connected",
        ERROR: "error"
    });
    const endpoints = Object.freeze({
        connections: "/ssh/api/connections",
        secrets: "/keys",
        revealSecret: "/keys/reveal",
        terminal: "/ssh/ws"
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
        savedConnection: document.getElementById("savedConnection"),
        closeModalButton: document.getElementById("closeModalBtn"),
        host: document.getElementById("host"),
        port: document.getElementById("port"),
        username: document.getElementById("username"),
        passwordSource: document.getElementById("passwordSource"),
        manualPasswordSource: document.getElementById("manual-password-source"),
        secretPasswordSource: document.getElementById("secret-password-source"),
        password: document.getElementById("password"),
        privateKey: document.getElementById("privateKey"),
        passphrase: document.getElementById("passphrase"),
        secretName: document.getElementById("secretName"),
        secretHelp: document.getElementById("secretHelp"),
        secretPassphrase: document.getElementById("secretPassphrase"),
        secretPassphraseField: document.getElementById("secretPassphraseField"),
        refreshSecretsButton: document.getElementById("refreshSecretsBtn"),
        saveConnection: document.getElementById("saveConnection"),
        connectionName: document.getElementById("connectionName"),
        connectionNameField: document.getElementById("connectionNameField"),
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
    let secrets = [];
    let savedConnections = [];

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
        term.write("Terminal - Click Connect to start an SSH session.\r\n");
    }

    function activeSocket(candidate) {
        return socket === candidate;
    }

    function websocketURL(path) {
        const url = new URL(path, window.location.href);
        url.protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        return url.toString();
    }

    function sendSocket(candidate, event, data = null) {
        if (!candidate || candidate.readyState !== WebSocket.OPEN) {
            return false;
        }
        candidate.send(JSON.stringify({ event, data }));
        return true;
    }

    function closeSocket(candidate = socket) {
        if (!candidate || !activeSocket(candidate)) {
            return;
        }
        sendSocket(candidate, "disconnect");
        socket = null;
        candidate.close(1000, "client disconnect");
    }

    function fitTerminal() {
        requestAnimationFrame(() => {
            fitAddon.fit();
            if (state === State.CONNECTED && socket) {
                sendSocket(socket, "resize", { cols: term.cols, rows: term.rows });
            }
        });
    }

    async function fetchJSON(url, options = {}) {
        const response = await fetch(url, {
            credentials: "same-origin",
            cache: "no-store",
            ...options,
            headers: {
                "Content-Type": "application/json",
                ...(options.headers || {})
            }
        });
        let payload;
        try {
            payload = await response.json();
        } catch {
            throw new Error(`Unexpected response from ${url}`);
        }
        if (!response.ok || payload.success === false) {
            throw new Error(payload.message || `Request failed with status ${response.status}`);
        }
        return payload;
    }

    function renderSavedConnections(selectedName = "") {
        elements.savedConnection.replaceChildren(new Option("New connection", ""));
        savedConnections.forEach(connection => {
            elements.savedConnection.add(new Option(connection.name, connection.name));
        });
        elements.savedConnection.value = selectedName;
    }

    async function loadSavedConnections(selectedName = elements.savedConnection.value) {
        try {
            const payload = await fetchJSON(endpoints.connections);
            savedConnections = Array.isArray(payload.connections) ? payload.connections : [];
            renderSavedConnections(selectedName);
        } catch (error) {
            savedConnections = [];
            renderSavedConnections();
            elements.savedConnection.add(new Option(`Unable to load: ${error.message}`, "", false, false));
        }
    }

    function selectedSecret() {
        return secrets.find(secret => secret.name === elements.secretName.value);
    }

    function updateSecretPassphrase() {
        const secret = selectedSecret();
        const needsPassphrase = secret?.encryption === "scrypt";
        elements.secretPassphraseField.hidden = !needsPassphrase;
        if (!needsPassphrase) {
            elements.secretPassphrase.value = "";
        }
        if (!secret) {
            elements.secretHelp.textContent = "Select a Secret Manager entry to use as the SSH password.";
            return;
        }
        const encryption = needsPassphrase ? "Passphrase required" : "No Secret Manager passphrase required";
        elements.secretHelp.textContent = secret.description
            ? `${secret.description} · ${encryption}`
            : encryption;
    }

    function renderSecrets(selectedName = "") {
        elements.secretName.replaceChildren(new Option("Select a secret", ""));
        secrets.forEach(secret => {
            const suffix = secret.encryption === "scrypt" ? " (passphrase)" : "";
            elements.secretName.add(new Option(`${secret.name}${suffix}`, secret.name));
        });
        elements.secretName.value = selectedName;
        updateSecretPassphrase();
    }

    async function loadSecrets(selectedName = elements.secretName.value) {
        elements.refreshSecretsButton.disabled = true;
        elements.secretName.disabled = true;
        elements.secretHelp.textContent = "Loading secrets…";
        try {
            const payload = await fetchJSON(endpoints.secrets);
            secrets = Array.isArray(payload.keys) ? payload.keys : [];
            renderSecrets(selectedName);
            if (selectedName && !selectedSecret()) {
                elements.secretHelp.textContent = `Saved secret "${selectedName}" is no longer available.`;
            } else if (secrets.length === 0) {
                elements.secretHelp.textContent = "No secrets are available in Secret Manager.";
            }
        } catch (error) {
            secrets = [];
            renderSecrets();
            elements.secretHelp.textContent = `Unable to load secrets: ${error.message}`;
        } finally {
            elements.refreshSecretsButton.disabled = false;
            elements.secretName.disabled = false;
        }
    }

    function setAuthMethod(nextMethod, loadSecretOptions = true) {
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
        if (loadSecretOptions && nextMethod === "password" && elements.passwordSource.value === "secret" && secrets.length === 0) {
            void loadSecrets();
        }
    }

    function setPasswordSource(source, loadSecretOptions = true) {
        const useSecret = source === "secret";
        elements.passwordSource.value = useSecret ? "secret" : "manual";
        elements.manualPasswordSource.hidden = useSecret;
        elements.secretPasswordSource.hidden = !useSecret;
        if (!useSecret) {
            elements.secretPassphrase.value = "";
        } else if (loadSecretOptions && secrets.length === 0) {
            void loadSecrets();
        }
    }

    async function applySavedConnection() {
        const connection = savedConnections.find(item => item.name === elements.savedConnection.value);
        if (!connection) {
            return;
        }
        elements.host.value = connection.host;
        elements.port.value = connection.port || "22";
        elements.username.value = connection.username;
        elements.privateKey.value = connection.private_key || "";
        elements.connectionName.value = connection.name;
        setAuthMethod(connection.auth_type, false);
        const usesSecretPassword = connection.auth_type === "password" && Boolean(connection.secret_name);
        setPasswordSource(usesSecretPassword ? "secret" : "manual", false);
        if (usesSecretPassword) {
            await loadSecrets(connection.secret_name || "");
        }
    }

    function toggleConnectionName() {
        elements.connectionNameField.hidden = !elements.saveConnection.checked;
        if (elements.saveConnection.checked && !elements.connectionName.value.trim()) {
            elements.connectionName.value = elements.savedConnection.value || elements.host.value.trim();
        }
    }

    function connectionSettingsFromForm() {
        if (!elements.saveConnection.checked) {
            return null;
        }
        const name = elements.connectionName.value.trim();
        if (!name) {
            throw new Error("Connection name is required when saving settings");
        }
        return {
            name,
            host: elements.host.value.trim(),
            port: elements.port.value.trim(),
            username: elements.username.value.trim(),
            auth_type: authMethod,
            private_key: authMethod === "key" ? elements.privateKey.value.trim() : "",
            secret_name: authMethod === "password" && elements.passwordSource.value === "secret"
                ? elements.secretName.value
                : ""
        };
    }

    async function saveConnectionSettings(settings) {
        if (!settings) {
            return;
        }
        await fetchJSON(endpoints.connections, {
            method: "POST",
            body: JSON.stringify(settings)
        });
        await loadSavedConnections(settings.name);
    }

    async function requestFromForm() {
        const host = elements.host.value.trim();
        const port = elements.port.value.trim();
        const username = elements.username.value.trim();
        if (!host || !port || !username) {
            throw new Error("Host, port, and username are required");
        }

        const request = { host, port, username };
        if (authMethod === "password") {
            if (elements.passwordSource.value === "manual") {
                request.password = elements.password.value;
            } else {
                const secret = selectedSecret();
                if (!secret) {
                    throw new Error("Select a Secret Manager entry");
                }
                const passphrase = elements.secretPassphrase.value;
                if (secret.encryption === "scrypt" && !passphrase) {
                    throw new Error("This secret requires a passphrase");
                }
                const revealed = await fetchJSON(endpoints.revealSecret, {
                    method: "POST",
                    body: JSON.stringify({
                        name: secret.name,
                        passphrase
                    })
                });
                if (typeof revealed.value !== "string" || !revealed.value) {
                    throw new Error("The selected secret is empty");
                }
                request.password = revealed.value;
            }
        } else if (authMethod === "key") {
            request.privateKey = elements.privateKey.value.trim();
            request.passphrase = elements.passphrase.value;
        }
        return request;
    }

    async function handleConnect() {
        if (state === State.CONNECTING || state === State.CONNECTED) {
            return;
        }

        let request;
        let settings;
        setState(State.CONNECTING, "Preparing connection…");
        try {
            settings = connectionSettingsFromForm();
            request = await requestFromForm();
        } catch (error) {
            setState(State.ERROR, error.message);
            return;
        }

        setState(State.CONNECTING, "Connecting…");
        const nextSocket = new WebSocket(websocketURL(endpoints.terminal));
        socket = nextSocket;

        nextSocket.addEventListener("open", () => {
            if (activeSocket(nextSocket)) {
                sendSocket(nextSocket, "connect_ssh", request);
                request.password = "";
                request.passphrase = "";
                elements.password.value = "";
                elements.passphrase.value = "";
                elements.secretPassphrase.value = "";
            }
        });

        nextSocket.addEventListener("message", event => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            let message;
            try {
                message = JSON.parse(event.data);
            } catch {
                setState(State.ERROR, "Error: Invalid response from SSH service");
                closeSocket(nextSocket);
                return;
            }
            switch (message.event) {
            case "ssh_connected": {
                const data = message.data || {};
                setState(State.CONNECTED, `Connected to ${data.user}@${data.host}:${data.port}`);
                elements.modal.close();
                term.clear();
                term.write(`Connected to ${data.user}@${data.host}:${data.port}\r\n`);
                term.focus();
                fitTerminal();
                if (settings) {
                    void saveConnectionSettings(settings).then(() => {
                        if (activeSocket(nextSocket)) {
                            setState(State.CONNECTED, `Connected to ${data.user}@${data.host}:${data.port} · Saved`);
                        }
                    }).catch(error => {
                        if (activeSocket(nextSocket)) {
                            setState(State.CONNECTED, `Connected · Unable to save settings: ${error.message}`);
                        }
                    });
                }
                break;
            }
            case "ssh_error":
                setState(State.ERROR, `Error: ${message.data || "SSH service error"}`);
                closeSocket(nextSocket);
                break;
            case "ssh_disconnected": {
                const reason = message.data || "SSH session closed";
                setState(State.IDLE, reason);
                term.write(`\r\n${reason}\r\n`);
                closeSocket(nextSocket);
                break;
            }
            case "terminal_output":
                term.write(message.data || "");
                break;
            default:
                break;
            }
        });

        nextSocket.addEventListener("error", () => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            setState(State.ERROR, "Connection to SSH service failed");
            closeSocket(nextSocket);
        });

        nextSocket.addEventListener("close", event => {
            if (!activeSocket(nextSocket)) {
                return;
            }
            socket = null;
            if (state === State.CONNECTED || state === State.CONNECTING) {
                const reason = event.reason ? `: ${event.reason}` : "";
                setState(State.IDLE, `Disconnected${reason}`);
                term.write("\r\nSSH session closed\r\n");
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
            void loadSavedConnections();
        }
    }

    elements.connectButton.addEventListener("click", toggleConnection);
    elements.closeModalButton.addEventListener("click", () => elements.modal.close());
    elements.connectionForm.addEventListener("submit", event => {
        event.preventDefault();
        void handleConnect();
    });
    elements.authTabs.forEach(tab => tab.addEventListener("click", () => setAuthMethod(tab.dataset.auth)));
    elements.passwordSource.addEventListener("change", () => setPasswordSource(elements.passwordSource.value));
    elements.savedConnection.addEventListener("change", () => void applySavedConnection());
    elements.secretName.addEventListener("change", updateSecretPassphrase);
    elements.refreshSecretsButton.addEventListener("click", () => void loadSecrets());
    elements.saveConnection.addEventListener("change", toggleConnectionName);
    elements.modal.addEventListener("click", event => {
        if (event.target === elements.modal) {
            elements.modal.close();
        }
    });
    term.onData(data => {
        if (state === State.CONNECTED && socket) {
            sendSocket(socket, "terminal_input", data);
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
    setPasswordSource(elements.passwordSource.value, false);
    void loadSavedConnections();
    writeWelcome();
    fitTerminal();
    term.focus();
})();
