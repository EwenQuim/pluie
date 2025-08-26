// Tree folding functionality
function toggleFolder(folderPath) {
	const folderElement = document.getElementById('folder-' + folderPath);
	const chevronElement = document.getElementById('chevron-' + folderPath);

	if (folderElement && chevronElement) {
		const isCurrentlyOpen = folderElement.style.display !== 'none';

		if (isCurrentlyOpen) {
			// Close folder
			folderElement.style.display = 'none';
			chevronElement.textContent = '▶';
		} else {
			// Open folder
			folderElement.style.display = 'block';
			chevronElement.textContent = '▼';
		}

		// Store folder state in localStorage
		const openFolders = JSON.parse(localStorage.getItem('openFolders') || '{}');
		openFolders[folderPath] = !isCurrentlyOpen;
		localStorage.setItem('openFolders', JSON.stringify(openFolders));
	}
}

// Expand all folders
function expandAllFolders() {
	const allFolders = document.querySelectorAll('[id^="folder-"]');
	const allChevrons = document.querySelectorAll('[id^="chevron-"]');
	const openFolders = {};

	allFolders.forEach(folder => {
		folder.style.display = 'block';
		const folderPath = folder.id.replace('folder-', '');
		openFolders[folderPath] = true;
	});

	allChevrons.forEach(chevron => {
		chevron.textContent = '▼';
	});

	// Update localStorage
	localStorage.setItem('openFolders', JSON.stringify(openFolders));
}

// Collapse all folders
function collapseAllFolders() {
	const allFolders = document.querySelectorAll('[id^="folder-"]');
	const allChevrons = document.querySelectorAll('[id^="chevron-"]');
	const openFolders = {};

	allFolders.forEach(folder => {
		folder.style.display = 'none';
		const folderPath = folder.id.replace('folder-', '');
		openFolders[folderPath] = false;
	});

	allChevrons.forEach(chevron => {
		chevron.textContent = '▶';
	});

	// Update localStorage
	localStorage.setItem('openFolders', JSON.stringify(openFolders));
}

// Restore folder states from localStorage
function restoreFolderStates() {
	const openFolders = JSON.parse(localStorage.getItem('openFolders') || '{}');

	for (const [folderPath, isOpen] of Object.entries(openFolders)) {
		const folderElement = document.getElementById('folder-' + folderPath);
		const chevronElement = document.getElementById('chevron-' + folderPath);

		if (folderElement && chevronElement) {
			if (isOpen) {
				folderElement.style.display = 'block';
				chevronElement.textContent = '▼';
			} else {
				folderElement.style.display = 'none';
				chevronElement.textContent = '▶';
			}
		}
	}
}

// Keyboard shortcuts
document.addEventListener('DOMContentLoaded', function () {
	// Restore folder states when page loads
	restoreFolderStates();

	// Handle cmd+K (or ctrl+K on Windows/Linux) to focus search
	document.addEventListener('keydown', function (event) {
		// Check for cmd+K on Mac or ctrl+K on Windows/Linux
		if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
			event.preventDefault();

			// Find the search input
			const searchInput = document.querySelector('input[name="search"]');
			if (searchInput) {
				searchInput.focus();
				searchInput.select(); // Select all text for easy replacement
			}
		}
	});

	// Restore folder states after HTMX requests
	document.body.addEventListener('htmx:afterSwap', function (event) {
		// Only restore states if the notes list was updated
		if (event.target.id === 'notes-list' || event.target.closest('#notes-list')) {
			setTimeout(restoreFolderStates, 10); // Small delay to ensure DOM is updated
		}
	});
});
