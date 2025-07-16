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

// Re-initialize after HTMX content loads
document.addEventListener('htmx:afterSwap', (event) => {
    console.log('HTMX afterSwap event:', event.target.id, event.target.className);
    
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

// Gallery Slideshow functionality
let currentSlideIndex = 0;
let galleryImages = [];

function openGallerySlideshow(galleryId) {
    // Collect all images from the gallery
    const galleryCard = document.querySelector(`[data-gallery-id="${galleryId}"]`);
    if (!galleryCard) return;
    
    const galleryTitle = galleryCard.querySelector('h3').textContent;
    galleryImages = [];
    
    // First, try to get all images from gallery assets data (this contains ALL images)
    const mediaData = window.galleryAssetsData || {};
    let hasAssetsData = false;
    
    if (mediaData[galleryId] && mediaData[galleryId].Data) {
        hasAssetsData = true;
        mediaData[galleryId].Data.forEach(asset => {
            if (asset.FileInfo && asset.FileInfo.Url) {
                galleryImages.push({
                    url: asset.FileInfo.Url,
                    alt: asset.AltText || '',
                    title: asset.Title || '',
                    description: asset.Description || '',
                    credit: asset.Credit || ''
                });
            }
        });
    }
    
    // If no assets data available, fall back to gallery preview images (limited)
    if (!hasAssetsData) {
        const previewImages = galleryCard.querySelectorAll('.gallery-preview-image');
        previewImages.forEach(img => {
            galleryImages.push({
                url: img.src,
                alt: img.alt,
                title: img.title || '',
                description: '',
                credit: ''
            });
        });
    }
    
    if (galleryImages.length === 0) return;
    
    // Setup modal
    const modal = document.getElementById('gallery-slideshow-modal');
    const titleElement = document.getElementById('slideshow-title');
    
    titleElement.textContent = galleryTitle;
    currentSlideIndex = 0;
    
    // Show modal
    modal.style.display = 'block';
    document.body.style.overflow = 'hidden';
    
    // Setup slideshow
    updateSlide();
    createThumbnails();
    
    // Add keyboard navigation
    document.addEventListener('keydown', handleKeyNavigation);
}

function closeGallerySlideshow(event) {
    if (event && event.target !== event.currentTarget && !event.target.classList.contains('slideshow-close')) {
        return;
    }
    
    const modal = document.getElementById('gallery-slideshow-modal');
    modal.style.display = 'none';
    document.body.style.overflow = 'auto';
    
    // Remove keyboard navigation
    document.removeEventListener('keydown', handleKeyNavigation);
}

function previousSlide() {
    currentSlideIndex = currentSlideIndex > 0 ? currentSlideIndex - 1 : galleryImages.length - 1;
    updateSlide();
}

function nextSlide() {
    currentSlideIndex = currentSlideIndex < galleryImages.length - 1 ? currentSlideIndex + 1 : 0;
    updateSlide();
}

function goToSlide(index) {
    currentSlideIndex = index;
    updateSlide();
}

function updateSlide() {
    if (galleryImages.length === 0) return;
    
    const image = galleryImages[currentSlideIndex];
    const slideImage = document.getElementById('slideshow-image');
    const slideTitle = document.getElementById('slideshow-image-title');
    const slideDescription = document.getElementById('slideshow-image-description');
    const slideCredit = document.getElementById('slideshow-image-credit');
    const slideCurrent = document.getElementById('slideshow-current');
    const slideTotal = document.getElementById('slideshow-total');
    
    slideImage.src = image.url;
    slideImage.alt = image.alt;
    slideTitle.textContent = image.title;
    slideDescription.textContent = image.description;
    slideCredit.textContent = image.credit ? `Credit: ${image.credit}` : '';
    slideCurrent.textContent = currentSlideIndex + 1;
    slideTotal.textContent = galleryImages.length;
    
    // Update thumbnails
    updateThumbnailsActive();
}

function createThumbnails() {
    const thumbnailsContainer = document.getElementById('slideshow-thumbnails');
    thumbnailsContainer.innerHTML = '';
    
    galleryImages.forEach((image, index) => {
        const thumbnail = document.createElement('img');
        thumbnail.src = image.url;
        thumbnail.alt = image.alt;
        thumbnail.className = 'slideshow-thumbnail';
        thumbnail.onclick = () => goToSlide(index);
        
        if (index === currentSlideIndex) {
            thumbnail.classList.add('active');
        }
        
        thumbnailsContainer.appendChild(thumbnail);
    });
}

function updateThumbnailsActive() {
    const thumbnails = document.querySelectorAll('.slideshow-thumbnail');
    thumbnails.forEach((thumb, index) => {
        thumb.classList.toggle('active', index === currentSlideIndex);
    });
}

function handleKeyNavigation(event) {
    switch(event.key) {
        case 'ArrowLeft':
            event.preventDefault();
            previousSlide();
            break;
        case 'ArrowRight':
            event.preventDefault();
            nextSlide();
            break;
        case 'Escape':
            event.preventDefault();
            closeGallerySlideshow();
            break;
    }
}

// Store gallery assets data globally for slideshow access
window.galleryAssetsData = {};
