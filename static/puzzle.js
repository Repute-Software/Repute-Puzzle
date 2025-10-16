// Puzzle game state
let gameState = {
    gridSize: 3,
    tiles: [],
    emptyIndex: 0,
    moves: 0,
    timeElapsed: 0,
    timeLimit: 0,
    timerMode: 'immediate',    // When to start timer: "immediate", "first_move", "countdown"
    countdownTime: 120,        // Time for countdown mode
    timerStarted: false,       // Track if timer has started (for first_move mode)
    timer: null,
    isComplete: false,
    imageUrl: '',
    testingMode: false,
    scrambleMoves: [],         // Store the scramble sequence
    solutionMoves: [],         // Reverse of scramble (the solution)
    scrambleMovesCount: 50,    // Number of moves to scramble
    autoSolveSpeed: 50         // Speed for auto-solve animation
};

// Initialize the puzzle when the page loads
document.addEventListener('DOMContentLoaded', function() {
    const container = document.getElementById('puzzle-container');
    if (!container) return;

    const gameContainer = document.querySelector('.game-container');
    if (!gameContainer) return;

    gameState.gridSize = parseInt(container.dataset.gridSize);
    gameState.imageUrl = container.dataset.imageUrl;
    gameState.timeLimit = parseInt(container.dataset.timeLimit);
    gameState.timerMode = container.dataset.timerMode || 'immediate';
    gameState.countdownTime = parseInt(container.dataset.countdownTime) || 120;
    gameState.testingMode = container.dataset.testingMode === 'true';
    gameState.scrambleMovesCount = parseInt(container.dataset.scrambleMoves);
    gameState.autoSolveSpeed = parseInt(container.dataset.autoSolveSpeed);

    // Apply dynamic colors from data attributes
    applyDynamicColors(gameContainer);

    // Add resize listener to adjust puzzle scaling
    window.addEventListener('resize', adjustPuzzleScaling);

    initializePuzzle();
    
    // Set up button listeners
    document.getElementById('shuffle-btn').addEventListener('click', initializePuzzle);
    document.getElementById('hint-btn').addEventListener('click', showHint);
    
    // Add auto-solve button if in testing mode
    if (gameState.testingMode) {
        const autoSolveBtn = document.getElementById('auto-solve-btn');
        if (autoSolveBtn) {
            autoSolveBtn.addEventListener('click', autoSolve);
        }
    }
});

// Initialize or reset the puzzle
function initializePuzzle() {
    gameState.moves = 0;
    gameState.timeElapsed = 0;
    gameState.isComplete = false;
    gameState.timerStarted = false;
    gameState.scrambleMoves = [];
    gameState.solutionMoves = [];
    
    updateMoveCounter();
    
    // Stop existing timer
    if (gameState.timer) {
        clearInterval(gameState.timer);
    }
    
    // Start timer based on timer mode
    if (gameState.timeLimit > 0 || gameState.timerMode === 'countdown') {
        if (gameState.timerMode === 'immediate') {
            startTimer();
        } else if (gameState.timerMode === 'countdown') {
            startCountdownTimer();
        }
        // For 'first_move' mode, timer will start on first tile click
    }
    
    // Shuffle the puzzle (starts from solved state and records moves)
    shufflePuzzle();
    
    // Render the puzzle
    renderPuzzle();
}

// Shuffle the puzzle with a solvable configuration by recording moves
function shufflePuzzle() {
    // Start from solved state
    const totalTiles = gameState.gridSize * gameState.gridSize;
    gameState.tiles = Array.from({length: totalTiles}, (_, i) => i);
    gameState.emptyIndex = totalTiles - 1;
    gameState.scrambleMoves = [];
    
    // Make N random valid moves and record them
    for (let i = 0; i < gameState.scrambleMovesCount; i++) {
        const validMoves = getValidMoves(gameState.emptyIndex);
        
        // Avoid immediate reversals for better scrambling
        let filteredMoves = validMoves;
        if (i > 0 && gameState.scrambleMoves.length > 0) {
            const lastMove = gameState.scrambleMoves[i - 1];
            filteredMoves = validMoves.filter(m => m !== lastMove.from);
        }
        
        // If all moves filtered out, use all valid moves
        if (filteredMoves.length === 0) {
            filteredMoves = validMoves;
        }
        
        const moveIndex = filteredMoves[Math.floor(Math.random() * filteredMoves.length)];
        
        // Record the move (from, to, tileValue)
        gameState.scrambleMoves.push({
            from: gameState.emptyIndex,
            to: moveIndex,
            tileValue: gameState.tiles[moveIndex]
        });
        
        // Execute the move
        swapTiles(gameState.emptyIndex, moveIndex);
        gameState.emptyIndex = moveIndex;
    }
    
    // Create solution by reversing the scramble
    gameState.solutionMoves = [...gameState.scrambleMoves].reverse();
    
    console.log(`Puzzle scrambled with ${gameState.scrambleMovesCount} moves. Solution has ${gameState.solutionMoves.length} steps.`);
}

// Get valid moves for a given position
function getValidMoves(index) {
    const moves = [];
    const row = Math.floor(index / gameState.gridSize);
    const col = index % gameState.gridSize;
    
    // Up
    if (row > 0) moves.push(index - gameState.gridSize);
    // Down
    if (row < gameState.gridSize - 1) moves.push(index + gameState.gridSize);
    // Left
    if (col > 0) moves.push(index - 1);
    // Right
    if (col < gameState.gridSize - 1) moves.push(index + 1);
    
    return moves;
}

// Swap two tiles
function swapTiles(index1, index2) {
    [gameState.tiles[index1], gameState.tiles[index2]] = 
    [gameState.tiles[index2], gameState.tiles[index1]];
}

// Render the puzzle on the screen
function renderPuzzle() {
    const container = document.getElementById('puzzle-container');
    container.innerHTML = '';
    
    // Get next correct move in testing mode
    const nextMove = gameState.testingMode ? getNextCorrectMove() : null;
    
    gameState.tiles.forEach((tileValue, index) => {
        const tile = document.createElement('div');
        tile.className = 'puzzle-tile';
        tile.dataset.index = index;
        tile.dataset.value = tileValue;
        
        if (tileValue === gameState.gridSize * gameState.gridSize - 1) {
            // Empty tile
            tile.classList.add('empty');
        } else {
            // Calculate background position
            const tileRow = Math.floor(tileValue / gameState.gridSize);
            const tileCol = tileValue % gameState.gridSize;
            const bgX = (tileCol * 100) / (gameState.gridSize - 1);
            const bgY = (tileRow * 100) / (gameState.gridSize - 1);
            
            tile.style.backgroundImage = `url(${gameState.imageUrl})`;
            tile.style.backgroundPosition = `${bgX}% ${bgY}%`;
            tile.style.backgroundSize = `${gameState.gridSize * 100}% ${gameState.gridSize * 100}%`;
            
            // Highlight the correct next move in testing mode
            if (gameState.testingMode && nextMove !== null && index === nextMove) {
                tile.classList.add('next-move');
            }
            
            // Add click handler
            tile.addEventListener('click', () => handleTileClick(index));
        }
        
        container.appendChild(tile);
    });
}

// Handle tile click
function handleTileClick(index) {
    if (gameState.isComplete) return;
    
    // Check if clicked tile is adjacent to empty tile
    const validMoves = getValidMoves(gameState.emptyIndex);
    
    if (validMoves.includes(index)) {
        // Start timer on first move if timer_mode is "first_move"
        if (gameState.timerMode === 'first_move' && !gameState.timerStarted && gameState.moves === 0) {
            if (gameState.timeLimit > 0) {
                startTimer();
            }
        }
        
        // Swap tiles
        swapTiles(gameState.emptyIndex, index);
        gameState.emptyIndex = index;
        gameState.moves++;
        
        updateMoveCounter();
        renderPuzzle();
        
        // Check if puzzle is solved
        if (isPuzzleSolved()) {
            handlePuzzleComplete();
        }
    }
}

// Check if puzzle is solved
function isPuzzleSolved() {
    return gameState.tiles.every((value, index) => value === index);
}

// Handle puzzle completion
function handlePuzzleComplete() {
    gameState.isComplete = true;
    
    // Stop timer
    if (gameState.timer) {
        clearInterval(gameState.timer);
    }
    
    // Show success modal
    showSuccessModal();
}

// Show success modal
function showSuccessModal() {
    const modal = document.getElementById('success-modal');
    const finalMoves = document.getElementById('final-moves');
    const finalTime = document.getElementById('final-time');
    const movesInput = document.getElementById('moves-input');
    const timeInput = document.getElementById('time-input');
    
    finalMoves.textContent = gameState.moves;
    finalTime.textContent = formatTime(gameState.timeElapsed);
    movesInput.value = gameState.moves;
    timeInput.value = gameState.timeElapsed;
    
    modal.classList.add('show');
}

// Update move counter
function updateMoveCounter() {
    document.getElementById('move-count').textContent = gameState.moves;
}

// Start timer (count up or down depending on mode)
function startTimer() {
    gameState.timerStarted = true;
    const timeDisplay = document.getElementById('time-display');
    
    gameState.timer = setInterval(() => {
        gameState.timeElapsed++;
        
        if (gameState.timeLimit > 0) {
            const remaining = gameState.timeLimit - gameState.timeElapsed;
            timeDisplay.textContent = formatTime(remaining);
            
            if (remaining <= 0) {
                clearInterval(gameState.timer);
                alert('Time is up! Try again.');
                initializePuzzle();
            }
        } else {
            timeDisplay.textContent = formatTime(gameState.timeElapsed);
        }
    }, 1000);
}

// Start countdown timer (counts down from countdown_time)
function startCountdownTimer() {
    gameState.timerStarted = true;
    const timeDisplay = document.getElementById('time-display');
    gameState.timeElapsed = 0;
    
    // Set initial display
    timeDisplay.textContent = formatTime(gameState.countdownTime);
    
    gameState.timer = setInterval(() => {
        gameState.timeElapsed++;
        const remaining = gameState.countdownTime - gameState.timeElapsed;
        timeDisplay.textContent = formatTime(remaining);
        
        if (remaining <= 0) {
            clearInterval(gameState.timer);
            alert('Time is up! Try again.');
            initializePuzzle();
        }
    }, 1000);
}

// Format time as MM:SS
function formatTime(seconds) {
    const mins = Math.floor(Math.abs(seconds) / 60);
    const secs = Math.abs(seconds) % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// Get the next correct move to solve the puzzle
function getNextCorrectMove() {
    if (!gameState.testingMode || gameState.solutionMoves.length === 0) {
        return null;
    }
    
    // The current number of moves made is our position in the solution
    const stepIndex = gameState.moves;
    
    if (stepIndex >= gameState.solutionMoves.length) {
        return null; // Already solved or past solution
    }
    
    // Get the next move from the solution sequence
    // When scrambling: empty moved FROM → TO
    // To reverse: we need to click the tile at FROM position
    const nextSolutionMove = gameState.solutionMoves[stepIndex];
    
    // Return the tile position that should be moved into the empty space
    return nextSolutionMove.from;
}

// Auto-solve the puzzle (for testing mode)
function autoSolve() {
    if (gameState.isComplete) return;
    
    const autoSolveBtn = document.getElementById('auto-solve-btn');
    if (autoSolveBtn) {
        autoSolveBtn.disabled = true;
        // Use translated text if available
        const solvingText = autoSolveBtn.getAttribute('data-text-solving') || '🤖 Solving...';
        autoSolveBtn.textContent = solvingText;
    }
    
    // Solve step by step with animation using the recorded solution
    function solveStep() {
        if (isPuzzleSolved()) {
            handlePuzzleComplete();
            return;
        }
        
        // Check if we have more solution moves available
        if (gameState.moves >= gameState.solutionMoves.length) {
            // Should not happen, but fallback
            console.log('Auto-solve: Reached end of solution moves, completing manually');
            completeManually();
            return;
        }
        
        // Get the next solution move
        const nextMove = getNextCorrectMove();
        
        if (nextMove !== null) {
            // Make the move
            handleTileClick(nextMove);
            
            // Continue solving with configured speed
            setTimeout(solveStep, gameState.autoSolveSpeed);
        } else {
            // Should not happen with recorded solution
            console.log('Auto-solve: No valid move found, completing manually');
            completeManually();
        }
    }
    
    solveStep();
}

// Manually complete the puzzle (fallback if solver fails)
function completeManually() {
    // Just set tiles to correct positions
    gameState.tiles = Array.from({length: gameState.gridSize * gameState.gridSize}, (_, i) => i);
    gameState.emptyIndex = gameState.gridSize * gameState.gridSize - 1;
    renderPuzzle();
    
    // Small delay then trigger completion
    setTimeout(() => {
        handlePuzzleComplete();
    }, 100);
}

// Show hint (briefly show the solution)
function showHint() {
    const tiles = document.querySelectorAll('.puzzle-tile');
    
    tiles.forEach((tile, index) => {
        const correctValue = index;
        if (parseInt(tile.dataset.value) === correctValue) {
            tile.classList.add('correct');
        }
    });
    
    setTimeout(() => {
        tiles.forEach(tile => tile.classList.remove('correct'));
    }, 2000);
}

// Apply dynamic colors from data attributes
function applyDynamicColors(container) {
    const root = document.documentElement;
    
    // Get color values from data attributes
    const tileColor = container.dataset.tileColor || '#667eea';
    const tileHoverColor = container.dataset.tileHoverColor || '#5568d3';
    const backgroundColor = container.dataset.backgroundColor || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
    const textColor = container.dataset.textColor || '#333333';
    const buttonTextColor = container.dataset.buttonTextColor || '#ffffff';
    const borderColor = container.dataset.borderColor || '#333333';
    
    // Apply CSS custom properties
    root.style.setProperty('--primary-color', tileColor);
    root.style.setProperty('--primary-hover', tileHoverColor);
    root.style.setProperty('--background-color', backgroundColor);
    root.style.setProperty('--text-color', textColor);
    root.style.setProperty('--button-text-color', buttonTextColor);
    root.style.setProperty('--border-color', borderColor);
    
    // Apply background color to body if it's a gradient or solid color
    if (backgroundColor.includes('gradient')) {
        document.body.style.background = backgroundColor;
    } else {
        document.body.style.background = backgroundColor;
    }
    
    // Apply text color to relevant elements
    const textElements = document.querySelectorAll('h1, h2, h3, p, .stat-label, .stat-value');
    textElements.forEach(el => {
        el.style.color = textColor;
    });
    
    // Apply button text color
    const buttons = document.querySelectorAll('.btn');
    buttons.forEach(btn => {
        btn.style.color = buttonTextColor;
    });
    
    // Apply border color to puzzle grid
    const puzzleGrid = document.querySelector('.puzzle-grid');
    if (puzzleGrid) {
        puzzleGrid.style.borderColor = borderColor;
    }
}

// Adjust puzzle scaling on window resize
function adjustPuzzleScaling() {
    // Re-render the puzzle to adjust background sizes
    if (gameState.tiles && gameState.tiles.length > 0) {
        renderPuzzle();
    }
}

