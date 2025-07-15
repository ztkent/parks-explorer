// User dropdown functionality
function toggleDropdown(event) {
    event.stopPropagation();
    const dropdown = event.target.closest('.user-dropdown');
    const dropdownContent = dropdown.querySelector('.dropdown-content');
    
    // Close any other open dropdowns
    document.querySelectorAll('.dropdown-content.show').forEach(d => {
        if (d !== dropdownContent) {
            d.classList.remove('show');
        }
    });
    
    // Toggle current dropdown
    dropdownContent.classList.toggle('show');
}

// Mobile menu functionality
function toggleMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    mobileNav.classList.toggle('show');
}

function closeMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    mobileNav.classList.remove('show');
}

// Enhanced UX functionality
document.addEventListener('DOMContentLoaded', () => {
    // Close mobile menu and dropdowns when clicking outside
    document.addEventListener('click', function(event) {
        const mobileNav = document.getElementById('mobile-nav');
        const menuBtn = document.querySelector('.mobile-menu-btn');
        
        if (mobileNav && !mobileNav.contains(event.target) && 
            menuBtn && !menuBtn.contains(event.target)) {
            mobileNav.classList.remove('show');
        }
        
        // Close user dropdown when clicking outside
        const userDropdown = event.target.closest('.user-dropdown');
        if (!userDropdown) {
            document.querySelectorAll('.dropdown-content.show').forEach(dropdown => {
                dropdown.classList.remove('show');
            });
        }
    });
    
    // Smooth scrolling for anchor links
    document.addEventListener('click', function(e) {
        if (e.target.tagName === 'A' && e.target.getAttribute('href') && 
            e.target.getAttribute('href').startsWith('#')) {
            e.preventDefault();
            const targetId = e.target.getAttribute('href').substring(1);
            const targetElement = document.getElementById(targetId);
            
            if (targetElement) {
                targetElement.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
                closeMobileMenu();
            }
        }
    });

    // Initialize park detail tabs if on park detail page
    initializeParkTabs();
});

// Park detail page tab functionality
function initializeParkTabs() {
    const tabs = document.querySelectorAll('.park-nav-tab');
    
    if (tabs.length === 0) return; // Not on park detail page
    
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            // Remove active class from all tabs
            tabs.forEach(t => t.classList.remove('active'));
            
            // Add active class to clicked tab
            tab.classList.add('active');
            
            // Update URL hash without page reload
            const targetTab = tab.dataset.tab;
            if (history.pushState && targetTab) {
                const newUrl = window.location.pathname + '#' + targetTab;
                history.pushState(null, '', newUrl);
            }
        });
    });
    
    // Handle URL hash on page load
    const hash = window.location.hash.substr(1);
    if (hash) {
        const targetTab = document.querySelector(`[data-tab="${hash}"]`);
        
        if (targetTab) {
            // Remove active from all tabs
            tabs.forEach(t => t.classList.remove('active'));
            
            // Activate target tab and trigger its HTMX request
            targetTab.classList.add('active');
            htmx.trigger(targetTab, 'click');
        }
    }
}

// Infinite scrolling functionality
let currentOffset = 12; // Start after featured parks (first 12)
let isLoading = false;
let hasMoreParks = true;
const parksPerPage = 12;

// Initialize infinite scrolling after HTMX loads featured parks
document.addEventListener('htmx:afterSwap', (event) => {
    if (event.target.id === 'parksGrid') {
        // Remove any duplicate parks that may have been added
        removeDuplicateParks();
        
        // Show the infinite scroll trigger after featured parks load
        const trigger = document.getElementById('infinite-scroll-trigger');
        const searchInput = document.querySelector('.hero-search-input');
        
        // If search input is empty (showing featured parks), reset and show trigger
        if (searchInput && searchInput.value.trim() === '' && trigger && hasMoreParks) {
            currentOffset = 12;
            hasMoreParks = true;
            trigger.style.display = 'flex';
            setupInfiniteScrollTrigger();
        } else if (searchInput && searchInput.value.trim() !== '' && trigger) {
            // If there's a search query, hide the infinite scroll trigger
            trigger.style.display = 'none';
        }
    }
});

// Setup the infinite scroll trigger
function setupInfiniteScrollTrigger() {
    const trigger = document.getElementById('infinite-scroll-trigger');
    if (!trigger || !hasMoreParks) return;
    
    // Set up the HTMX attributes for the next load
    trigger.setAttribute('hx-get', `/api/parks?offset=${currentOffset}&limit=${parksPerPage}`);
    trigger.setAttribute('hx-trigger', 'intersect once');
    trigger.setAttribute('hx-target', '#infinite-scroll-trigger');
    trigger.setAttribute('hx-swap', 'none');
    
    // Process the element so HTMX recognizes it
    htmx.process(trigger);
}

// Handle infinite scroll responses
document.addEventListener('htmx:beforeRequest', (event) => {
    if (event.target.id === 'infinite-scroll-trigger') {
        isLoading = true;
    }
});

document.addEventListener('htmx:afterRequest', (event) => {
    if (event.target.id === 'infinite-scroll-trigger') {
        isLoading = false;
        
        const response = event.detail.xhr.responseText;
        
        // Check if we got park cards or if we're at the end
        if (!response.includes('park-card') || response.includes('No more parks')) {
            hasMoreParks = false;
            event.target.style.display = 'none';
        } else {
            // Parse the response and check for duplicates
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = response;
            const newParkCards = tempDiv.querySelectorAll('.park-card');
            const existingParksGrid = document.getElementById('parksGrid');
            const existingParkSlugs = new Set();
            
            // Get all existing park slugs
            existingParksGrid.querySelectorAll('.park-card').forEach(card => {
                const slug = card.dataset.park;
                if (slug) {
                    existingParkSlugs.add(slug);
                }
            });
            
            // Filter out duplicate parks and add only unique ones
            let addedNewParks = false;
            newParkCards.forEach(card => {
                const slug = card.dataset.park;
                if (slug && !existingParkSlugs.has(slug)) {
                    existingParksGrid.appendChild(card.cloneNode(true));
                    addedNewParks = true;
                }
            });
            
            // If no new unique parks were added, consider it the end
            if (!addedNewParks) {
                hasMoreParks = false;
                event.target.style.display = 'none';
            } else {
                // Update offset and setup trigger for next load
                currentOffset += parksPerPage;
                setTimeout(() => {
                    setupInfiniteScrollTrigger();
                }, 100);
            }
        }
    }
});

// Reset infinite scroll state when search is performed
document.addEventListener('input', (event) => {
    if (event.target.classList.contains('hero-search-input')) {
        // Reset infinite scroll state when search starts
        const trigger = document.getElementById('infinite-scroll-trigger');
        if (trigger) {
            trigger.style.display = 'none';
        }
        isLoading = false;
        hasMoreParks = true;
        currentOffset = 12;
    }
});

// Function to remove duplicate park cards based on data-park attribute
function removeDuplicateParks() {
    const parksGrid = document.getElementById('parksGrid');
    if (!parksGrid) return;
    
    const parkCards = parksGrid.querySelectorAll('.park-card[data-park]');
    const seenParks = new Set();
    
    parkCards.forEach(card => {
        const parkSlug = card.getAttribute('data-park');
        if (seenParks.has(parkSlug)) {
            card.remove();
        } else {
            seenParks.add(parkSlug);
        }
    });
}

// Expandable cards functionality for park details
function initializeExpandableCards() {
    const expandableCards = document.querySelectorAll('.expandable-card');
    
    expandableCards.forEach(card => {
        const expandButton = card.querySelector('.expand-button');
        const content = card.querySelector('.card-content');
        
        if (!expandButton || !content) return;
        
        // Force a reflow to get accurate measurements
        card.style.display = 'block';
        
        // Temporarily remove collapsed class to measure full height
        const wasCollapsed = card.classList.contains('collapsed');
        card.classList.remove('collapsed');
        
        // Get the actual content height
        const actualHeight = content.scrollHeight;
        
        // Restore collapsed state
        if (wasCollapsed) {
            card.classList.add('collapsed');
        }
        
        // Get the collapsed height from CSS
        const styles = getComputedStyle(card);
        const collapsedHeight = parseInt(styles.maxHeight) || 300;
                
        // If content is shorter than collapsed height, hide the expand button
        if (actualHeight <= collapsedHeight + 50) { // 50px buffer
            expandButton.classList.add('hidden');
            card.classList.remove('collapsed');
            return;
        }
        
        // Ensure the expand button is visible and the card is collapsed
        expandButton.classList.remove('hidden');
        card.classList.add('collapsed');
        
        // Remove any existing event listeners to avoid duplicates
        const newButton = expandButton.cloneNode(true);
        expandButton.parentNode.replaceChild(newButton, expandButton);
        
        // Add click handler for expand/collapse
        newButton.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            
            const isCollapsed = card.classList.contains('collapsed');
            
            if (isCollapsed) {
                // Expand
                card.classList.remove('collapsed');
                newButton.innerHTML = 'Show Less <span class="expand-icon">▼</span>';
            } else {
                // Collapse
                card.classList.add('collapsed');
                newButton.innerHTML = 'Show More <span class="expand-icon">▼</span>';
                
                // Scroll to top of card when collapsing
                card.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }
        });
    });
}

// Debug function to test expandable cards
function debugExpandableCards() {
    console.log('=== Debugging Expandable Cards ===');
    const cards = document.querySelectorAll('.expandable-card');
    console.log(`Found ${cards.length} expandable cards`);
    
    cards.forEach((card, index) => {
        const button = card.querySelector('.expand-button');
        const content = card.querySelector('.card-content');
        console.log(`Card ${index}:`, {
            hasButton: !!button,
            hasContent: !!content,
            isCollapsed: card.classList.contains('collapsed'),
            cardHeight: card.offsetHeight,
            contentHeight: content ? content.scrollHeight : 'N/A',
            maxHeight: getComputedStyle(card).maxHeight
        });
    });
}

// Re-initialize expandable cards after HTMX content loads
document.addEventListener('htmx:afterSwap', (event) => {
    // If we're on a park details page and content was swapped
    if (event.target.id === 'dynamic-content' || 
        event.target.closest('.park-details-grid')) {
        setTimeout(() => {
            console.log('Re-initializing expandable cards after HTMX swap');
            initializeExpandableCards();
        }, 200); // Slightly longer delay to ensure content is fully rendered
    }
    
    if (event.target.id === 'parksGrid') {
        // Remove any duplicate parks that may have been added
        removeDuplicateParks();
        
        // Show the infinite scroll trigger after featured parks load
        const trigger = document.getElementById('infinite-scroll-trigger');
        const searchInput = document.querySelector('.hero-search-input');
        
        // If search input is empty (showing featured parks), reset and show trigger
        if (searchInput && searchInput.value.trim() === '' && trigger && hasMoreParks) {
            currentOffset = 12;
            hasMoreParks = true;
            trigger.style.display = 'flex';
            setupInfiniteScrollTrigger();
        } else if (searchInput && searchInput.value.trim() !== '' && trigger) {
            // If there's a search query, hide the infinite scroll trigger
            trigger.style.display = 'none';
        }
    }
});
