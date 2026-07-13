(() => {
    "use strict";

    const storageKeys = Object.freeze({
        theme: "arupa.theme",
        language: "arupa.language"
    });
    const listeners = {
        theme: new Set(),
        language: new Set()
    };
    const storageListeners = new Set();

    function read(key, fallback) {
        try {
            return window.localStorage.getItem(key) || fallback;
        } catch (_) {
            return fallback;
        }
    }

    function write(key, value) {
        try {
            window.localStorage.setItem(key, value);
        } catch (_) {
            // Keep the in-memory event API usable when storage is unavailable.
        }
    }

    function notify(name, value, source) {
        listeners[name].forEach((listener) => {
            try {
                listener(value, { name, source });
            } catch (error) {
                setTimeout(() => { throw error; });
            }
        });
    }

    function notifyStorage(key, value, previous, source) {
        storageListeners.forEach((listener) => {
            try {
                listener({ key, value, previous, source });
            } catch (error) {
                setTimeout(() => { throw error; });
            }
        });
    }

    function normalizeTheme(theme) {
        return theme === "dark" ? "dark" : "light";
    }

    function normalizeLanguage(language) {
        return String(language || "").trim().toLowerCase();
    }

    function setValue(name, value, normalize) {
        const normalized = normalize(value);
        const previous = read(storageKeys[name], "");
        write(storageKeys[name], normalized);
        if (previous !== normalized) notify(name, normalized, "local");
        return normalized;
    }

    function subscribe(name, listener) {
        if (typeof listener !== "function") return () => {};
        listeners[name].add(listener);
        return () => listeners[name].delete(listener);
    }

    window.addEventListener("storage", (event) => {
        if (event.storageArea !== window.localStorage) return;
        notifyStorage(event.key, event.newValue, event.oldValue, "storage");
        if (event.key === storageKeys.theme) {
            notify("theme", normalizeTheme(event.newValue), "storage");
        }
        if (event.key === storageKeys.language) {
            notify("language", normalizeLanguage(event.newValue), "storage");
        }
    });

    const sdk = {
        storageKeys,
        getStorage(key, fallback = null) {
            return read(String(key), fallback);
        },
        setStorage(key, value) {
            const normalizedKey = String(key);
            const previous = read(normalizedKey, null);
            const next = String(value);
            write(normalizedKey, next);
            if (previous !== next) notifyStorage(normalizedKey, next, previous, "local");
            return previous !== next;
        },
        removeStorage(key) {
            const normalizedKey = String(key);
            const previous = read(normalizedKey, null);
            try {
                window.localStorage.removeItem(normalizedKey);
            } catch (_) {
                // Ignore storage access failures.
            }
            if (previous !== null) notifyStorage(normalizedKey, null, previous, "local");
        },
        onStorageChange(listener) {
            if (typeof listener !== "function") return () => {};
            storageListeners.add(listener);
            return () => storageListeners.delete(listener);
        },
        getTheme() {
            return normalizeTheme(read(storageKeys.theme, "light"));
        },
        setTheme(theme) {
            const previous = this.getTheme();
            const normalized = setValue("theme", theme, normalizeTheme);
            if (previous !== normalized) notifyStorage(storageKeys.theme, normalized, previous, "local");
            return normalized;
        },
        onThemeChange(listener) {
            return subscribe("theme", listener);
        },
        getLanguage() {
            return normalizeLanguage(read(storageKeys.language, ""));
        },
        setLanguage(language) {
            const previous = this.getLanguage();
            const normalized = setValue("language", language, normalizeLanguage);
            if (previous !== normalized) notifyStorage(storageKeys.language, normalized, previous, "local");
            return normalized;
        },
        onLanguageChange(listener) {
            return subscribe("language", listener);
        }
    };

    window.webSDK = Object.freeze(sdk);
    window.WebSDK = window.webSDK;
})();
