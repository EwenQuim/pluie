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

// Handle TOC click events for smooth scrolling and active state
function handleTOCClick(event, tocLink) {
	event.preventDefault();

	const targetId = tocLink.getAttribute('href').substring(1); // Remove the #
	const targetElement = document.getElementById(targetId);

	if (targetElement) {
		// Smooth scroll to target
		targetElement.scrollIntoView({
			behavior: 'smooth',
			block: 'start'
		});

		// Update URL hash
		history.pushState(null, null, `#${targetId}`);

		// Update active state
		updateActiveTocItem(tocLink);
	}
}

// Slugify function to match the unified Go implementation for headings
function slugify(text) {
	return text
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');
}

// Add IDs to headings that match the TOC structure
function addHeadingIds() {
	const contentArea = document.querySelector('.prose');
	if (!contentArea) return;

	const headings = contentArea.querySelectorAll('h1, h2, h3, h4, h5, h6');
	const usedSlugs = new Map(); // Track used slugs to handle duplicates

	headings.forEach((heading) => {
		if (!heading.id) {
			const text = heading.textContent.trim();
			const baseSlug = slugify(text);
			let id = baseSlug;

			// Handle duplicate slugs by appending a number
			if (usedSlugs.has(baseSlug)) {
				const count = usedSlugs.get(baseSlug) + 1;
				usedSlugs.set(baseSlug, count);
				id = `${baseSlug}-${count}`;
			} else {
				usedSlugs.set(baseSlug, 0);
			}

			heading.id = id;
		}
	});
}

// Update active TOC item based on scroll position
function updateActiveTocItem(activeItem = null) {
	const tocLinks = document.querySelectorAll('#table-of-contents a');

	// Remove active class from all items
	tocLinks.forEach(link => {
		link.classList.remove('text-purple-600', 'bg-purple-50', 'font-medium');
		link.classList.add('text-gray-600');
	});

	if (activeItem) {
		// Set specific item as active
		activeItem.classList.remove('text-gray-600');
		activeItem.classList.add('text-purple-600', 'bg-purple-50', 'font-medium');
	} else {
		// Find active item based on scroll position
		const headings = document.querySelectorAll('.prose h1, .prose h2, .prose h3, .prose h4, .prose h5, .prose h6');
		let activeHeading = null;

		// Find the heading that's currently in view
		headings.forEach(heading => {
			const rect = heading.getBoundingClientRect();
			if (rect.top <= 100 && rect.top >= -100) {
				activeHeading = heading;
			}
		});

		if (activeHeading) {
			const activeLink = document.querySelector(`#table-of-contents a[href="#${activeHeading.id}"]`);
			if (activeLink) {
				activeLink.classList.remove('text-gray-600');
				activeLink.classList.add('text-purple-600', 'bg-purple-50', 'font-medium');
			}
		}
	}
}

// Throttle function for scroll events
function throttle(func, wait) {
	let timeout;
	return function executedFunction(...args) {
		const later = () => {
			clearTimeout(timeout);
			func(...args);
		};
		clearTimeout(timeout);
		timeout = setTimeout(later, wait);
	};
}

// Mobile sidebar functionality
function toggleMobileSidebar() {
	const sidebar = document.getElementById('mobile-sidebar');
	const overlay = document.getElementById('mobile-sidebar-overlay');
	const line1 = document.getElementById('burger-line-1');
	const line2 = document.getElementById('burger-line-2');
	const line3 = document.getElementById('burger-line-3');

	if (sidebar && overlay) {
		const isHidden = sidebar.classList.contains('-translate-x-full');

		if (isHidden) {
			// Show sidebar
			sidebar.classList.remove('-translate-x-full');
			sidebar.classList.add('translate-x-0');
			overlay.classList.remove('opacity-0', 'invisible');
			overlay.classList.add('opacity-100', 'visible');

			// Animate burger menu to X
			if (line1 && line2 && line3) {
				line1.classList.add('rotate-45', 'translate-y-2');
				line2.classList.add('opacity-0');
				line3.classList.add('-rotate-45', '-translate-y-2');
			}

			// Prevent body scroll when sidebar is open
			document.body.style.overflow = 'hidden';
		} else {
			// Hide sidebar
			closeMobileSidebar();
		}
	}
}

function closeMobileSidebar() {
	const sidebar = document.getElementById('mobile-sidebar');
	const overlay = document.getElementById('mobile-sidebar-overlay');
	const line1 = document.getElementById('burger-line-1');
	const line2 = document.getElementById('burger-line-2');
	const line3 = document.getElementById('burger-line-3');

	if (sidebar && overlay) {
		sidebar.classList.remove('translate-x-0');
		sidebar.classList.add('-translate-x-full');
		overlay.classList.remove('opacity-100', 'visible');
		overlay.classList.add('opacity-0', 'invisible');

		// Reset burger menu animation
		if (line1 && line2 && line3) {
			line1.classList.remove('rotate-45', 'translate-y-2');
			line2.classList.remove('opacity-0');
			line3.classList.remove('-rotate-45', '-translate-y-2');
		}

		// Restore body scroll
		document.body.style.overflow = '';
	}
}

// Close mobile sidebar when clicking on a link (for better UX)
function handleMobileLinkClick() {
	// Only close on mobile devices
	if (window.innerWidth < 768) {
		closeMobileSidebar();
	}
}

// YAML front matter toggle functionality
function toggleYamlFrontmatter() {
	const yamlContent = document.getElementById('yaml-content');
	const yamlToggleBtn = document.getElementById('yaml-toggle-btn');

	if (yamlContent && yamlToggleBtn) {
		const isCurrentlyHidden = yamlContent.style.display === 'none';

		if (isCurrentlyHidden) {
			// Show YAML content
			yamlContent.style.display = 'block';
			yamlToggleBtn.querySelector('span:last-child').textContent = 'Hide';
		} else {
			// Hide YAML content
			yamlContent.style.display = 'none';
			yamlToggleBtn.querySelector('span:last-child').textContent = 'Show';
		}

		// Store YAML front matter state in localStorage
		localStorage.setItem('yamlFrontmatterOpen', isCurrentlyHidden.toString());
	}
}

// Restore YAML front matter state from localStorage
function restoreYamlFrontmatterState() {
	const yamlContent = document.getElementById('yaml-content');
	const yamlToggleBtn = document.getElementById('yaml-toggle-btn');

	if (yamlContent && yamlToggleBtn) {
		const isOpen = localStorage.getItem('yamlFrontmatterOpen') === 'true';

		if (isOpen) {
			yamlContent.style.display = 'block';
			yamlToggleBtn.querySelector('span:last-child').textContent = 'Hide';
		} else {
			yamlContent.style.display = 'none';
			yamlToggleBtn.querySelector('span:last-child').textContent = 'Show';
		}
	}
}

// Keyboard shortcuts
document.addEventListener('DOMContentLoaded', function () {
	// Restore folder states when page loads
	restoreFolderStates();

	// Restore YAML front matter state when page loads
	restoreYamlFrontmatterState();

	// Add IDs to headings to match TOC structure
	addHeadingIds();

	// Update active TOC item on scroll
	const mainContent = document.querySelector('.flex-1.overflow-y-auto');
	if (mainContent) {
		mainContent.addEventListener('scroll', throttle(updateActiveTocItem, 100));
	}

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

	// Add heading IDs after HTMX content updates
	document.body.addEventListener('htmx:afterSwap', function (event) {
		// Check if the main content was updated
		if (event.target.closest('.prose') || event.target.classList.contains('prose')) {
			setTimeout(() => {
				addHeadingIds();
				// Handle hash in URL on page load
				if (window.location.hash) {
					const targetElement = document.querySelector(window.location.hash);
					if (targetElement) {
						setTimeout(() => {
							targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
							const activeLink = document.querySelector(`#table-of-contents a[href="${window.location.hash}"]`);
							if (activeLink) {
								updateActiveTocItem(activeLink);
							}
						}, 100);
					}
				}
			}, 50);
		}
	});

	// Handle hash in URL on initial page load
	if (window.location.hash) {
		setTimeout(() => {
			const targetElement = document.querySelector(window.location.hash);
			if (targetElement) {
				targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
				const activeLink = document.querySelector(`#table-of-contents a[href="${window.location.hash}"]`);
				if (activeLink) {
					updateActiveTocItem(activeLink);
				}
			}
		}, 100);
	}
});
