document.getElementById('gameForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const gameID = document.getElementById('gameID').value.trim();
    if (gameID) {
        updateGame(gameID);
    } else {
        createGame();
    }
});

function createGame() {
    saveGame('POST', '/games');
}

function updateGame(gameID) {
    saveGame('PUT', `/games/${gameID}`);
}

function saveGame(method, url) {
    try {
        const gameData = getGameData();
        fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(gameData)
        })
        .then(response => handleResponse(response))
        .then(data => {
            document.getElementById('gameForm').reset();
            document.getElementById('errorMessage').style.display = 'none';
            listGames();
        })
        .catch(error => showError(error));
    } catch (error) {
        showError(error);
    }
}

function getGameData() {
    const gameID = document.getElementById('gameID').value.trim();
    const gameName = document.getElementById('gameName').value.trim();
    if (!gameName) {
        throw new Error('Game name is required');
    }
    const data = {
        game: gameName,
        description: document.getElementById('description').value.trim(),
        status: document.getElementById('status').value
    };
    // Include game_id only if provided (for updates)
    if (gameID) {
        data.game_id = gameID;
    }
    return data;
}

function handleResponse(response) {
    if (!response.ok) {
        throw new Error(`Error: ${response.statusText}`);
    }
    return response.json();
}

function showError(error) {
    const errorMessage = document.getElementById('errorMessage');
    errorMessage.textContent = error.message;
    errorMessage.style.display = 'block';
}

function listGames() {
    fetch('/games')
        .then(response => handleResponse(response))
        .then(games => {
            const gameList = document.getElementById('gameList');
            gameList.innerHTML = '';
            games.forEach(game => {
                const gameItem = document.createElement('div');
                gameItem.className = 'game-item';

                // Game name and ID
                const strong = document.createElement('strong');
                strong.textContent = game.game;
                const gameIdSpan = document.createElement('span');
                gameIdSpan.textContent = ` (ID: ${game.game_id})`;
                
                // Description
                const descPara = document.createElement('p');
                descPara.textContent = `Description: ${game.description}`;

                // Status
                const statusPara = document.createElement('p');
                statusPara.textContent = `Status: ${game.status}`;

                // Edit Button
                const editButton = document.createElement('button');
                editButton.className = 'edit-button';
                editButton.textContent = 'Edit';
                editButton.addEventListener('click', () => editGame(game.game_id, game.game, game.description, game.status));

                // Delete Button
                const deleteButton = document.createElement('button');
                deleteButton.className = 'delete-button';
                deleteButton.textContent = 'Delete';
                deleteButton.addEventListener('click', () => deleteGame(game.game_id));

                // Assemble elements
                gameItem.appendChild(strong);
                gameItem.appendChild(gameIdSpan);
                gameItem.appendChild(descPara);
                gameItem.appendChild(statusPara);
                gameItem.appendChild(editButton);
                gameItem.appendChild(deleteButton);

                gameList.appendChild(gameItem);
            });
        })
        .catch(error => showError(error));
}

function editGame(gameID, gameName, description, status) {
    const gameIDField = document.getElementById('gameID');
    gameIDField.value = gameID;
    gameIDField.readOnly = true; // Prevent ID modification during edit
    document.getElementById('gameName').value = gameName;
    document.getElementById('description').value = description;
    document.getElementById('status').value = status;
}

function deleteGame(gameID) {
    if (confirm('Are you sure you want to delete this game?')) {
        fetch(`/games/${gameID}`, { method: 'DELETE' })
            .then(handleResponse)
            .then(() => listGames())
            .catch(error => showError(error));
    }
}

// Initialize game list on page load
document.addEventListener('DOMContentLoaded', listGames);