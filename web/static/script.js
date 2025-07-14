// Mobile menu functionality
function toggleMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    mobileNav.classList.toggle('show');
}

function closeMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    mobileNav.classList.remove('show');
}

// Minimal JavaScript for enhanced UX
document.addEventListener('DOMContentLoaded', () => {
    // Close mobile menu when clicking outside
    document.addEventListener('click', function(event) {
        const mobileNav = document.getElementById('mobile-nav');
        const menuBtn = document.querySelector('.mobile-menu-btn');
        
        if (mobileNav && !mobileNav.contains(event.target) && 
            menuBtn && !menuBtn.contains(event.target)) {
            mobileNav.classList.remove('show');
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
                
                // Close mobile menu if open
                closeMobileMenu();
            }
        }
    });
});

// Handle login dropdown interactions after HTMX loads content
document.addEventListener('htmx:afterSwap', (event) => {
    // Check if auth container was updated and has a dropdown
    if (event.target.classList.contains('auth-container') || event.target.closest('.auth-container')) {
        const userAvatar = document.getElementById('userAvatar');
        const userDropdown = document.getElementById('userDropdown');
        if (userAvatar && userDropdown) {
            userAvatar.addEventListener('click', (e) => {
                e.stopPropagation();
                userDropdown.classList.toggle('show');
            });
            // Close dropdown when clicking outside
            document.addEventListener('click', () => {
                userDropdown.classList.remove('show');
            });
        }
    }
});

// Park detail page tab functionality
document.addEventListener('DOMContentLoaded', () => {
    // Initialize park detail tabs if on park detail page
    initializeParkTabs();
});

function initializeParkTabs() {
    const tabs = document.querySelectorAll('.park-nav-tab');
    const tabContents = document.querySelectorAll('.park-tab-content');
    
    if (tabs.length === 0) return; // Not on park detail page
    
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const targetTab = tab.dataset.tab;
            
            // Remove active class from all tabs and content
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Add active class to clicked tab and corresponding content
            tab.classList.add('active');
            const targetContent = document.getElementById(targetTab);
            if (targetContent) {
                targetContent.classList.add('active');
            }
            
            // Update URL hash without page reload
            if (history.pushState) {
                const newUrl = window.location.pathname + '#' + targetTab;
                history.pushState(null, '', newUrl);
            }
        });
    });
    
    // Handle URL hash on page load
    const hash = window.location.hash.substr(1);
    if (hash) {
        const targetTab = document.querySelector(`[data-tab="${hash}"]`);
        const targetContent = document.getElementById(hash);
        
        if (targetTab && targetContent) {
            // Remove active from all
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Activate target
            targetTab.classList.add('active');
            targetContent.classList.add('active');
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
        // Show the infinite scroll trigger after featured parks load
        const trigger = document.getElementById('infinite-scroll-trigger');
        if (trigger && hasMoreParks) {
            trigger.style.display = 'flex';
            setupInfiniteScrollTrigger();
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
    trigger.setAttribute('hx-target', '#parksGrid');
    trigger.setAttribute('hx-swap', 'beforeend');
    
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
            // Update offset and setup trigger for next load
            currentOffset += parksPerPage;
            setTimeout(() => {
                setupInfiniteScrollTrigger();
            }, 100); // Small delay to ensure DOM is updated
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

// Handle search completion and show infinite scroll trigger again when search is cleared
document.addEventListener('htmx:afterSwap', (event) => {
    if (event.target.id === 'parksGrid') {
        const searchInput = document.querySelector('.hero-search-input');
        const trigger = document.getElementById('infinite-scroll-trigger');
        
        // If search input is empty or has whitespace only (showing featured parks), reset and show trigger
        if (searchInput && searchInput.value.trim() === '' && trigger) {
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
