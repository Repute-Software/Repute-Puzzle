// Embed-specific functionality for iframe integration

// Check if we're running in an iframe
const isInIframe = window.parent !== window;

// Send postMessage events to parent window
function sendEmbedEvent(eventType, data = {}) {
    if (isInIframe) {
        window.parent.postMessage({
            source: 'repute-puzzle',
            type: eventType,
            data: data,
            timestamp: Date.now()
        }, '*');
    }
}

// Apply custom primary color from URL parameter
function applyCustomColor() {
    const container = document.querySelector('[data-embed="true"]');
    if (!container) return;

    const primaryColor = container.dataset.primaryColor;
    if (primaryColor && /^#[0-9A-Fa-f]{6}$/.test(primaryColor)) {
        // Set CSS custom property
        document.documentElement.style.setProperty('--primary-color', primaryColor);
        
        // Calculate a slightly darker shade for hover (reduce brightness by 10%)
        const rgb = hexToRgb(primaryColor);
        if (rgb) {
            const darker = {
                r: Math.max(0, rgb.r - 25),
                g: Math.max(0, rgb.g - 25),
                b: Math.max(0, rgb.b - 25)
            };
            const hoverColor = rgbToHex(darker.r, darker.g, darker.b);
            document.documentElement.style.setProperty('--primary-hover', hoverColor);
        }
    }
}

// Convert hex color to RGB
function hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : null;
}

// Convert RGB to hex
function rgbToHex(r, g, b) {
    return "#" + [r, g, b].map(x => {
        const hex = x.toString(16);
        return hex.length === 1 ? "0" + hex : hex;
    }).join('');
}

// Initialize embed functionality
document.addEventListener('DOMContentLoaded', function() {
    // Only run embed code if we're in embed mode
    if (!document.querySelector('[data-embed="true"]')) {
        return;
    }

    // Apply custom color
    applyCustomColor();

    // Send loaded event
    sendEmbedEvent('puzzle.loaded', {
        gridSize: gameState.gridSize,
        timeLimit: gameState.timeLimit
    });

    // Track first move
    let firstMoveMade = false;
    const originalMoveTile = window.moveTile;
    window.moveTile = function(index) {
        const result = originalMoveTile(index);
        if (result && !firstMoveMade) {
            firstMoveMade = true;
            sendEmbedEvent('puzzle.started', {
                gridSize: gameState.gridSize
            });
        }
        return result;
    };

    // Track puzzle completion
    const originalCheckWin = window.checkWin;
    window.checkWin = function() {
        const wasComplete = gameState.isComplete;
        originalCheckWin();
        
        if (!wasComplete && gameState.isComplete) {
            sendEmbedEvent('puzzle.completed', {
                moves: gameState.moves,
                time: gameState.timeElapsed,
                gridSize: gameState.gridSize
            });
        }
    };

    // Track new puzzle
    const originalShufflePuzzle = window.shufflePuzzle;
    window.shufflePuzzle = function() {
        originalShufflePuzzle();
        firstMoveMade = false;
        sendEmbedEvent('puzzle.reset', {
            gridSize: gameState.gridSize
        });
    };

    // Listen for successful discount code generation
    // We'll hook into HTMX events
    document.body.addEventListener('htmx:afterSwap', function(event) {
        if (event.detail.target.id === 'email-form') {
            // Check if it's a success response (contains discount code)
            const discountCodeElement = document.querySelector('.discount-code-display');
            if (discountCodeElement) {
                const discountCode = discountCodeElement.textContent.trim();
                sendEmbedEvent('discount.generated', {
                    code: discountCode,
                    moves: gameState.moves,
                    time: gameState.timeElapsed
                });
            }
        }
    });
});

