// @ts-check
// Constants
const SCROLL_HEADING_THRESHOLD = 100;
const SCROLL_THROTTLE_MS = 100;
const HTMX_RESTORE_DELAY_MS = 10;
const HTMX_HEADING_DELAY_MS = 50;
const HASH_SCROLL_DELAY_MS = 100;
const MOBILE_BREAKPOINT = 768;

// Helper functions for localStorage
/**
 * Retrieves the open folders state from localStorage.
 * @returns {Record<string, boolean>} Object mapping folder paths to their open state (true/false)
 */
function getOpenFolders() {
	return JSON.parse(localStorage.getItem('openFolders') || '{}');
}

/**
 * Saves the open folders state to localStorage.
 * @param {Record<string, boolean>} openFolders - Object mapping folder paths to their open state (true/false)
 */
function saveOpenFolders(openFolders) {
	localStorage.setItem('openFolders', JSON.stringify(openFolders));
}

// Tree folding functionality
/**
 * Sets the visual state (open/closed) of a folder in the navigation tree.
 * @param {string} folderPath - The path identifier of the folder
 * @param {boolean} isOpen - Whether the folder should be open (true) or closed (false)
 */
function setFolderState(folderPath, isOpen) {
	const folderElement = document.getElementById('folder-' + folderPath);
	const chevronElement = document.getElementById('chevron-' + folderPath);

	if (folderElement && chevronElement) {
		folderElement.style.display = isOpen ? 'block' : 'none';
		chevronElement.textContent = isOpen ? '▼' : '▶';
	}
}

/**
 * Toggles a folder between open and closed states and persists the state to localStorage.
 * @param {string} folderPath - The path identifier of the folder to toggle
 */
function toggleFolder(folderPath) {
	const folderElement = document.getElementById('folder-' + folderPath);

	if (folderElement) {
		const isCurrentlyOpen = folderElement.style.display !== 'none';
		const newState = !isCurrentlyOpen;

		setFolderState(folderPath, newState);

		// Store folder state in localStorage
		const openFolders = getOpenFolders();
		openFolders[folderPath] = newState;
		saveOpenFolders(openFolders);
	}
}

/**
 * Sets all folders in the navigation tree to the same state (all open or all closed).
 * @param {boolean} isOpen - Whether all folders should be open (true) or closed (false)
 */
function setAllFoldersState(isOpen) {
	const allFolders = document.querySelectorAll('[id^="folder-"]');
	/** @type {Record<string, boolean>} */
	const openFolders = {};

	allFolders.forEach(folder => {
		const folderPath = folder.id.replace('folder-', '');
		setFolderState(folderPath, isOpen);
		openFolders[folderPath] = isOpen;
	});

	// Update localStorage
	saveOpenFolders(openFolders);
}

/**
 * Expands all folders in the navigation tree.
 */
function expandAllFolders() {
	setAllFoldersState(true);
}

/**
 * Collapses all folders in the navigation tree.
 */
function collapseAllFolders() {
	setAllFoldersState(false);
}

/**
 * Restores folder states from localStorage, setting each folder to its previously saved state.
 */
function restoreFolderStates() {
	const openFolders = getOpenFolders();

	for (const [folderPath, isOpen] of Object.entries(openFolders)) {
		setFolderState(folderPath, isOpen);
	}
}

/**
 * Handles clicks on table of contents links, providing smooth scrolling and updating active state.
 * @param {Event} event - The click event
 * @param {HTMLAnchorElement} tocLink - The clicked TOC link element
 */
function handleTOCClick(event, tocLink) {
	event.preventDefault();

	const href = tocLink.getAttribute('href');
	if (!href) return;

	const targetId = href.substring(1); // Remove the #
	const targetElement = document.getElementById(targetId);

	if (targetElement) {
		// Smooth scroll to target
		targetElement.scrollIntoView({
			behavior: 'smooth',
			block: 'start'
		});

		// Update URL hash
		history.pushState(null, '', `#${targetId}`);

		// Update active state
		updateActiveTocItem(tocLink);
	}
}

/**
 * Converts text to a URL-friendly slug format, matching the Go implementation.
 * Lowercases text, replaces non-alphanumeric characters with hyphens, and trims edge hyphens.
 * @param {string} text - The text to slugify
 * @returns {string} The slugified text
 */
function slugify(text) {
	return text
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');
}

/**
 * Adds unique IDs to all headings in the content area to match the TOC structure.
 * Handles duplicate slugs by appending numbers.
 */
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

/**
 * Updates the active state of TOC items based on scroll position or a specific item.
 * @param {Element|null} activeItem - The specific TOC link to set as active, or null to determine from scroll position
 */
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
		/** @type {Element|null} */
		let activeHeading = null;

		// Find the heading that's currently in view
		headings.forEach(heading => {
			const rect = heading.getBoundingClientRect();
			if (rect.top <= SCROLL_HEADING_THRESHOLD && rect.top >= -SCROLL_HEADING_THRESHOLD) {
				activeHeading = heading;
			}
		});

		if (activeHeading) {
			const headingId = /** @type {Element} */ (activeHeading).id;
			const activeLink = document.querySelector(`#table-of-contents a[href="#${headingId}"]`);
			if (activeLink) {
				activeLink.classList.remove('text-gray-600');
				activeLink.classList.add('text-purple-600', 'bg-purple-50', 'font-medium');
			}
		}
	}
}

/**
 * Handles URL hash navigation by smoothly scrolling to the target heading and updating TOC active state.
 * Checks if there's a hash in the URL and scrolls to the corresponding heading if found.
 */
function handleHashNavigation() {
	if (window.location.hash) {
		const targetElement = document.querySelector(window.location.hash);
		if (targetElement) {
			targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
			const activeLink = document.querySelector(`#table-of-contents a[href="${window.location.hash}"]`);
			if (activeLink) {
				updateActiveTocItem(activeLink);
			}
		}
	}
}

/**
 * Creates a debounced version of a function that delays its execution until after
 * the specified wait time has elapsed since the last invocation.
 * @param {(...args: any[]) => void} func - The function to debounce
 * @param {number} wait - The delay in milliseconds
 * @returns {(...args: any[]) => void} The debounced function
 */
function debounce(func, wait) {
	/** @type {number | undefined} */
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
/**
 * Retrieves all mobile sidebar-related DOM elements.
 * @returns {{sidebar: HTMLElement|null, overlay: HTMLElement|null, line1: HTMLElement|null, line2: HTMLElement|null, line3: HTMLElement|null}} Object containing sidebar, overlay, and burger menu line elements
 */
function getMobileSidebarElements() {
	return {
		sidebar: document.getElementById('mobile-sidebar'),
		overlay: document.getElementById('mobile-sidebar-overlay'),
		line1: document.getElementById('burger-line-1'),
		line2: document.getElementById('burger-line-2'),
		line3: document.getElementById('burger-line-3')
	};
}

/**
 * Sets the visual state of the burger menu icon (hamburger or X).
 * @param {boolean} isOpen - Whether to show the menu as open (X) or closed (hamburger)
 */
function setBurgerMenuState(isOpen) {
	const { line1, line2, line3 } = getMobileSidebarElements();

	if (line1 && line2 && line3) {
		if (isOpen) {
			line1.classList.add('rotate-45', 'translate-y-2');
			line2.classList.add('opacity-0');
			line3.classList.add('-rotate-45', '-translate-y-2');
		} else {
			line1.classList.remove('rotate-45', 'translate-y-2');
			line2.classList.remove('opacity-0');
			line3.classList.remove('-rotate-45', '-translate-y-2');
		}
	}
}

/**
 * Toggles the mobile sidebar between open and closed states.
 * Also animates the burger menu icon and manages body scroll locking.
 */
function toggleMobileSidebar() {
	const { sidebar, overlay } = getMobileSidebarElements();

	if (sidebar && overlay) {
		const isHidden = sidebar.classList.contains('-translate-x-full');

		if (isHidden) {
			// Show sidebar
			sidebar.classList.remove('-translate-x-full');
			sidebar.classList.add('translate-x-0');
			overlay.classList.remove('opacity-0', 'invisible');
			overlay.classList.add('opacity-100', 'visible');

			// Animate burger menu to X
			setBurgerMenuState(true);

			// Prevent body scroll when sidebar is open
			document.body.style.overflow = 'hidden';
		} else {
			// Hide sidebar
			closeMobileSidebar();
		}
	}
}

/**
 * Closes the mobile sidebar and restores the page to its normal state.
 * Resets the burger menu icon and restores body scrolling.
 */
function closeMobileSidebar() {
	const { sidebar, overlay } = getMobileSidebarElements();

	if (sidebar && overlay) {
		sidebar.classList.remove('translate-x-0');
		sidebar.classList.add('-translate-x-full');
		overlay.classList.remove('opacity-100', 'visible');
		overlay.classList.add('opacity-0', 'invisible');

		// Reset burger menu animation
		setBurgerMenuState(false);

		// Restore body scroll
		document.body.style.overflow = '';
	}
}

/**
 * Closes the mobile sidebar when a link is clicked, but only on mobile devices.
 * Provides better UX by automatically hiding the sidebar after navigation.
 */
function handleMobileLinkClick() {
	// Only close on mobile devices
	if (window.innerWidth < MOBILE_BREAKPOINT) {
		closeMobileSidebar();
	}
}

// YAML front matter toggle functionality
/**
 * Toggles the visibility of YAML front matter and persists the state to localStorage.
 * Updates the button text to reflect the current state (Show/Hide).
 */
function toggleYamlFrontmatter() {
	const yamlContent = document.getElementById('yaml-content');
	const yamlToggleBtn = document.getElementById('yaml-toggle-btn');

	if (yamlContent && yamlToggleBtn) {
		const isCurrentlyHidden = yamlContent.style.display === 'none';
		const buttonText = yamlToggleBtn.querySelector('span:last-child');

		if (isCurrentlyHidden) {
			// Show YAML content
			yamlContent.style.display = 'block';
			if (buttonText) buttonText.textContent = 'Hide';
		} else {
			// Hide YAML content
			yamlContent.style.display = 'none';
			if (buttonText) buttonText.textContent = 'Show';
		}

		// Store YAML front matter state in localStorage
		localStorage.setItem('yamlFrontmatterOpen', isCurrentlyHidden.toString());
	}
}

/**
 * Restores the YAML front matter visibility state from localStorage on page load.
 * Sets the initial visibility and button text based on saved preferences.
 */
function restoreYamlFrontmatterState() {
	const yamlContent = document.getElementById('yaml-content');
	const yamlToggleBtn = document.getElementById('yaml-toggle-btn');

	if (yamlContent && yamlToggleBtn) {
		const isOpen = localStorage.getItem('yamlFrontmatterOpen') === 'true';
		const buttonText = yamlToggleBtn.querySelector('span:last-child');

		if (isOpen) {
			yamlContent.style.display = 'block';
			if (buttonText) buttonText.textContent = 'Hide';
		} else {
			yamlContent.style.display = 'none';
			if (buttonText) buttonText.textContent = 'Show';
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
		mainContent.addEventListener('scroll', debounce(updateActiveTocItem, SCROLL_THROTTLE_MS));
	}

	// Handle cmd+K (or ctrl+K on Windows/Linux) to focus search
	document.addEventListener('keydown', function (event) {
		// Check for cmd+K on Mac or ctrl+K on Windows/Linux
		if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
			event.preventDefault();

			// Find the search input
			const searchInput = /** @type {HTMLInputElement|null} */ (document.querySelector('input[name="search"]'));
			if (searchInput) {
				searchInput.focus();
				searchInput.select(); // Select all text for easy replacement
			}
		}
	});

	// Restore folder states after HTMX requests
	document.body.addEventListener('htmx:afterSwap', function (event) {
		// Only restore states if the notes list was updated
		const target = /** @type {Element|null} */ (event.target);
		if (target && (target.id === 'notes-list' || target.closest('#notes-list'))) {
			setTimeout(restoreFolderStates, HTMX_RESTORE_DELAY_MS); // Small delay to ensure DOM is updated
		}
	});

	// Add heading IDs after HTMX content updates
	document.body.addEventListener('htmx:afterSwap', function (event) {
		// Check if the main content was updated
		const target = /** @type {Element|null} */ (event.target);
		if (target && (target.closest('.prose') || target.classList.contains('prose'))) {
			setTimeout(() => {
				addHeadingIds();
				// Handle hash in URL on page load
				setTimeout(handleHashNavigation, HASH_SCROLL_DELAY_MS);
			}, HTMX_HEADING_DELAY_MS);
		}
	});

	// Handle hash in URL on initial page load
	setTimeout(handleHashNavigation, HASH_SCROLL_DELAY_MS);
});
