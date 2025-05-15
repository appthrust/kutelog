// Browser and OS detection logic
function detectBrowser() {
	const ua = navigator.userAgent;

	if (ua.includes("Safari") && !ua.includes("Chrome")) {
		return "Safari";
	}
	if (ua.includes("Firefox")) {
		return "Firefox";
	}
	if (ua.includes("Edg")) {
		return "Edge";
	}
	if (ua.includes("Chrome")) {
		return "Chrome";
	}
	return "Other";
}

function detectOS() {
	return navigator.platform.toLowerCase().includes("mac") ? "mac" : "windows";
}

// Shortcut key definitions
const shortcuts = {
	Chrome: {
		mac: ["⌘", "⌥", "J"],
		windows: ["Ctrl", "Shift", "J"],
	},
	Firefox: {
		mac: ["⌘", "⌥", "K"],
		windows: ["Ctrl", "Shift", "K"],
	},
	Safari: {
		mac: ["⌘", "⌥", "C"],
	},
	Edge: {
		mac: ["⌘", "⌥", "J"],
		windows: ["Ctrl", "Shift", "J"],
	},
};

// Generate HTML for shortcut keys
function getShortcutHTML(browser, os) {
	const shortcut = shortcuts[browser]?.[os];
	if (!shortcut) return "";

	return shortcut
		.map((key) => `<kbd class="bg-gray-600 px-2 py-1 rounded">${key}</kbd>`)
		.join(" + ");
}

// Set shortcut keys on page load
document.addEventListener("DOMContentLoaded", () => {
	const browser = detectBrowser();
	const os = detectOS();

	const shortcutContainer = document.getElementById("shortcut-keys");
	if (shortcutContainer) {
		shortcutContainer.innerHTML = getShortcutHTML(browser, os);
	}
});

let ws = null;
let reconnectAttempt = 0;
let lastReceivedTimestamp = 0;
const INITIAL_RETRY_DELAY = 1000; // Start with 1 second delay

// Connection status indicator
function updateConnectionStatus(isConnected) {
	const indicator = document.getElementById("conn-status-dot");
	const label = document.getElementById("conn-status-label");
	if (!indicator || !label) return;

	if (isConnected) {
		indicator.className =
			"w-3 h-3 rounded-full bg-green-500 animate-pulse-slow";
		label.textContent = "Connected";
	} else {
		indicator.className = "w-3 h-3 rounded-full bg-yellow-500";
		label.textContent = "Connecting...";
	}
}

const TIMESTAMP_STYLE = "color: #888";
const LEVEL_STYLES = {
	info: "color: #3b82f6",
	warning: "color: #f59e0b",
	error: "color: #ef4444",
	debug: "color: #10b981",
};
const MESSAGE_STYLE = "color: inherit";

// Log counter state
const logCounts = {
	log: 0,
	info: 0,
	warning: 0,
	error: 0,
	debug: 0,
};

// Update counter element
function updateCounter(type) {
	const elementId = type === "warning" ? "warn" : type; // convert 'warning' to 'warn'
	const element = document.getElementById(`${elementId}-count`);
	if (element) {
		element.textContent = logCounts[type];
		// Animation effect
		element.classList.add("scale-110", "transition-transform");
		setTimeout(() => {
			element.classList.remove("scale-110");
		}, 200);
	}
}

const logFn = {
	info: (...args) => {
		logCounts.info++;
		updateCounter("info");
		console.info(...args);
	},
	warning: (...args) => {
		logCounts.warning++;
		updateCounter("warning");
		console.warn(...args);
	},
	error: (...args) => {
		logCounts.error++;
		updateCounter("error");
		console.error(...args);
	},
	debug: (...args) => {
		logCounts.debug++;
		updateCounter("debug");
		console.debug(...args);
	},
};

function connect() {
	ws = new WebSocket(`ws://${location.host}/ws`);

	ws.onopen = () => {
		console.log("Connected to WebSocket server");
		reconnectAttempt = 0;
		updateConnectionStatus(true);
	};

	ws.onmessage = (event) => {
		try {
			const message = JSON.parse(event.data);
			// Skip if message is already received or older
			if (message.id <= lastReceivedTimestamp) {
				return;
			}
			lastReceivedTimestamp = message.id;

			const data = message.body;
			if (
				typeof data === "object" &&
				data !== null &&
				typeof data.level === "string" &&
				typeof data.message === "string" &&
				typeof data.timestamp === "string"
			) {
				let formattedTimestamp;
				try {
					const timestamp = new Date(data.timestamp);
					if (Number.isNaN(timestamp.getTime())) {
						formattedTimestamp = data.timestamp;
					} else {
						formattedTimestamp = timestamp.toLocaleString();
					}
				} catch {
					formattedTimestamp = data.timestamp;
				}
				const msg = `%c${formattedTimestamp}%c ${data.level} %c${data.message}`;
				const args = [
					msg,
					TIMESTAMP_STYLE,
					LEVEL_STYLES[data.level] ?? "color: inherit",
					MESSAGE_STYLE,
				];
				if (data.data !== undefined) {
					args.push(data.data);
				}
				if (data.stack !== undefined) {
					args.push(data.stack);
				}
				const fn = logFn[data.level] ?? console.log;
				if (fn === console.log) {
					logCounts.log++;
					updateCounter("log");
				}
				fn(...args);
			} else if (typeof data === "string") {
				logCounts.log++;
				updateCounter("log");
				console.log(data);
			} else {
				logCounts.log++;
				updateCounter("log");
				console.log(message);
			}
		} catch {
			updateCounter("log");
			console.log(event.data);
		}
	};

	ws.onclose = () => {
		updateConnectionStatus(false);
		reconnect();
	};

	ws.onerror = () => {
		updateConnectionStatus(false);
	};
}

function reconnect() {
	const retryDelay = INITIAL_RETRY_DELAY;
	setTimeout(() => {
		reconnectAttempt++;
		connect();
	}, retryDelay);
}

// Set initial connection status
updateConnectionStatus(false);

// Initial connection
connect();
