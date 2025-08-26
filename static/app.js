// Keyboard shortcuts
document.addEventListener('DOMContentLoaded', function () {
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
});
