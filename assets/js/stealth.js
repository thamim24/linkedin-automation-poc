// Stealth JavaScript - Patches browser APIs to avoid detection
// This script is injected before page load to make the browser appear human

(function() {
    'use strict';

    // Remove webdriver flag
    Object.defineProperty(navigator, 'webdriver', {
        get: () => undefined,
        configurable: true
    });

    // Override plugins
    Object.defineProperty(navigator, 'plugins', {
        get: () => [
            {
                0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
                description: "Portable Document Format",
                filename: "internal-pdf-viewer",
                length: 1,
                name: "Chrome PDF Plugin"
            },
            {
                0: {type: "application/pdf", suffixes: "pdf", description: ""},
                description: "",
                filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
                length: 1,
                name: "Chrome PDF Viewer"
            },
            {
                0: {type: "application/x-nacl", suffixes: "", description: "Native Client Executable"},
                1: {type: "application/x-pnacl", suffixes: "", description: "Portable Native Client Executable"},
                description: "",
                filename: "internal-nacl-plugin",
                length: 2,
                name: "Native Client"
            }
        ],
        configurable: true
    });

    // Override mimeTypes
    Object.defineProperty(navigator, 'mimeTypes', {
        get: () => [
            {type: "application/pdf", suffixes: "pdf", description: ""},
            {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
            {type: "application/x-nacl", suffixes: "", description: "Native Client Executable"},
            {type: "application/x-pnacl", suffixes: "", description: "Portable Native Client Executable"}
        ],
        configurable: true
    });

    // Chrome runtime patch
    if (!window.chrome) {
        window.chrome = {};
    }
    
    if (!window.chrome.runtime) {
        window.chrome.runtime = {};
    }

    // Permissions API patch
    const originalQuery = window.navigator.permissions.query;
    window.navigator.permissions.query = (parameters) => (
        parameters.name === 'notifications' ?
            Promise.resolve({state: Notification.permission}) :
            originalQuery(parameters)
    );

    // Override notification permissions
    Object.defineProperty(Notification, 'permission', {
        get: () => 'default',
        configurable: true
    });

    // Battery API - make it unavailable (common in real browsers)
    if (navigator.getBattery) {
        navigator.getBattery = undefined;
    }

    // Connection API spoofing
    Object.defineProperty(navigator, 'connection', {
        get: () => ({
            effectiveType: '4g',
            downlink: 10,
            rtt: 50,
            saveData: false
        }),
        configurable: true
    });

    // Device memory spoofing
    Object.defineProperty(navigator, 'deviceMemory', {
        get: () => 8,
        configurable: true
    });

    // Override toString methods to hide modifications
    const makeNative = (fn, name) => {
        const handler = {
            apply: function(target, thisArg, args) {
                return fn.apply(thisArg, args);
            },
            get: function(target, prop) {
                if (prop === 'toString') {
                    return function() {
                        return `function ${name}() { [native code] }`;
                    };
                }
                return target[prop];
            }
        };
        return new Proxy(fn, handler);
    };

    // Patch navigator.permissions.query toString
    navigator.permissions.query = makeNative(navigator.permissions.query, 'query');

    // WebGL vendor spoofing
    const getParameter = WebGLRenderingContext.prototype.getParameter;
    WebGLRenderingContext.prototype.getParameter = function(parameter) {
        if (parameter === 37445) {
            return 'Intel Inc.';
        }
        if (parameter === 37446) {
            return 'Intel Iris OpenGL Engine';
        }
        return getParameter.apply(this, [parameter]);
    };

    // Canvas fingerprint protection (add slight noise)
    const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
    HTMLCanvasElement.prototype.toDataURL = function() {
        const context = this.getContext('2d');
        if (context) {
            const imageData = context.getImageData(0, 0, this.width, this.height);
            for (let i = 0; i < imageData.data.length; i += 4) {
                imageData.data[i] = imageData.data[i] + Math.random() * 0.1;
            }
            context.putImageData(imageData, 0, 0);
        }
        return originalToDataURL.apply(this, arguments);
    };

    // AudioContext fingerprint protection
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    if (AudioContext) {
        const originalCreateOscillator = AudioContext.prototype.createOscillator;
        AudioContext.prototype.createOscillator = function() {
            const oscillator = originalCreateOscillator.apply(this, arguments);
            const originalStart = oscillator.start;
            oscillator.start = function() {
                originalStart.apply(this, arguments);
            };
            return oscillator;
        };
    }

    // Screen properties (keep consistent with viewport)
    const screenProps = {
        availHeight: screen.height,
        availWidth: screen.width,
        availTop: 0,
        availLeft: 0,
        colorDepth: 24,
        pixelDepth: 24
    };

    Object.keys(screenProps).forEach(prop => {
        Object.defineProperty(screen, prop, {
            get: () => screenProps[prop],
            configurable: true
        });
    });

    // Date.prototype.getTimezoneOffset override
    // Will be set dynamically by Go code based on fingerprint

    // Remove automation markers from errors
    const originalError = Error;
    window.Error = function(...args) {
        const error = new originalError(...args);
        if (error.stack) {
            error.stack = error.stack.replace(/\bat (Object\.)?__webdriver[\w$]+/g, 'at Object.anonymous');
        }
        return error;
    };

    // Iframe contentWindow property
    const originalContentWindow = Object.getOwnPropertyDescriptor(HTMLIFrameElement.prototype, 'contentWindow');
    Object.defineProperty(HTMLIFrameElement.prototype, 'contentWindow', {
        get: function() {
            const win = originalContentWindow.get.call(this);
            if (win) {
                try {
                    win.navigator.webdriver = undefined;
                } catch (e) {}
            }
            return win;
        }
    });

    console.log('🎭 Stealth mode activated');
})();