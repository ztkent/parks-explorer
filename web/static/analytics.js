// Google Analytics tracking functions
class AnalyticsTracker {
    constructor() {
        this.config = null;
        this.initialized = false;
        this.init();
    }

    async init() {
        try {
            const response = await fetch('/api/analytics/config');
            this.config = await response.json();
            
            if (this.config.enabled && this.config.googleAnalyticsId) {
                this.loadGoogleAnalytics();
            } else if (this.config.debug) {
                console.log('Google Analytics disabled or no tracking ID provided');
            }
        } catch (error) {
            console.error('Failed to load analytics configuration:', error);
        }
    }

    loadGoogleAnalytics() {
        // Load gtag script
        const script = document.createElement('script');
        script.async = true;
        script.src = `https://www.googletagmanager.com/gtag/js?id=${this.config.googleAnalyticsId}`;
        document.head.appendChild(script);

        // Initialize gtag
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        window.gtag = gtag;
        
        gtag('js', new Date());
        gtag('config', this.config.googleAnalyticsId, {
            debug_mode: this.config.debug,
            send_page_view: true
        });

        this.initialized = true;

        if (this.config.debug) {
            console.log('Google Analytics initialized with ID:', this.config.googleAnalyticsId);
        }
    }

    // Track page views
    trackPageView(page_title, page_location) {
        if (!this.initialized || !this.config.enabled) {
            if (this.config && this.config.debug) {
                console.log('GA Debug: Page view tracked -', {page_title, page_location});
            }
            return;
        }

        gtag('config', this.config.googleAnalyticsId, {
            page_title: page_title,
            page_location: page_location
        });
    }

    // Track custom events
    trackEvent(action, category, label = null, value = null) {
        if (!this.initialized || !this.config.enabled) {
            if (this.config && this.config.debug) {
                console.log('GA Debug: Event tracked -', {action, category, label, value});
            }
            return;
        }

        const eventData = {
            event_category: category,
            event_label: label,
            value: value
        };

        // Remove null values
        Object.keys(eventData).forEach(key => 
            eventData[key] === null && delete eventData[key]
        );

        gtag('event', action, eventData);
    }

    // Track park searches
    trackParkSearch(searchTerm, resultCount) {
        this.trackEvent('search', 'park_search', searchTerm, resultCount);
    }

    // Track park page views
    trackParkView(parkName, parkCode) {
        this.trackEvent('view_item', 'park', `${parkName} (${parkCode})`);
    }

    // Track navigation
    trackNavigation(section) {
        this.trackEvent('navigation', 'menu_click', section);
    }

    // Track user engagement
    trackEngagement(action, details = null) {
        this.trackEvent(action, 'engagement', details);
    }

    // Track external link clicks
    trackExternalLink(url, context = null) {
        this.trackEvent('click', 'external_link', url);
    }

    // Track media interactions
    trackMediaInteraction(mediaType, action, mediaTitle = null) {
        this.trackEvent(action, `media_${mediaType}`, mediaTitle);
    }

    // Track search interactions
    trackSearchInteraction(searchType, query, filters = null) {
        let label = query;
        if (filters) {
            label += ` | ${Object.entries(filters).map(([k,v]) => `${k}:${v}`).join(', ')}`;
        }
        this.trackEvent('search', searchType, label);
    }

    // Track authentication events
    trackAuth(action) {
        this.trackEvent(action, 'authentication');
    }
}

// Initialize analytics tracker
const analytics = new AnalyticsTracker();

// Helper function to track page changes for HTMX
function trackHTMXPageChange(detail) {
    if (detail && detail.xhr && detail.xhr.responseURL) {
        const url = new URL(detail.xhr.responseURL);
        analytics.trackPageView(document.title, url.pathname);
    }
}

// Track HTMX requests
document.addEventListener('htmx:afterRequest', function(event) {
    if (event.detail.successful) {
        trackHTMXPageChange(event.detail);
    }
});

// Track clicks on park cards
document.addEventListener('click', function(event) {
    const parkCard = event.target.closest('.park-card');
    if (parkCard) {
        const parkName = parkCard.querySelector('.park-title')?.textContent;
        const parkSlug = parkCard.dataset.park;
        if (parkName && parkSlug) {
            analytics.trackParkView(parkName, parkSlug);
        }
    }
    
    // Track external links
    const link = event.target.closest('a[href^="http"]');
    if (link && !link.href.includes(window.location.hostname)) {
        analytics.trackExternalLink(link.href, link.textContent);
    }
    
    // Track navigation
    const navLink = event.target.closest('.nav-link');
    if (navLink) {
        analytics.trackNavigation(navLink.textContent);
    }
});

// Track search inputs
let searchTimeout;
document.addEventListener('input', function(event) {
    if (event.target.matches('input[type="text"]') && 
        (event.target.placeholder.includes('search') || event.target.name === 'q')) {
        
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            const query = event.target.value.trim();
            if (query.length > 2) {
                analytics.trackSearchInteraction('general_search', query);
            }
        }, 1000); // Debounce search tracking
    }
});

// Track form submissions for enhanced search tracking
document.addEventListener('submit', function(event) {
    const form = event.target.closest('form');
    if (form) {
        const formData = new FormData(form);
        const filters = {};
        let searchQuery = '';
        
        // Extract search query and filters
        for (let [key, value] of formData.entries()) {
            if (key === 'q' || key === 'activity-search') {
                searchQuery = value;
            } else if (value && value !== '') {
                filters[key] = value;
            }
        }
        
        // Determine search type based on form action or page
        let searchType = 'general_search';
        if (form.action.includes('/things-to-do/search')) {
            searchType = 'activities_search';
        } else if (form.action.includes('/events/search')) {
            searchType = 'events_search';
        } else if (form.action.includes('/camping/search')) {
            searchType = 'camping_search';
        } else if (form.action.includes('/news/search')) {
            searchType = 'news_search';
        }
        
        analytics.trackSearchInteraction(searchType, searchQuery, Object.keys(filters).length > 0 ? filters : null);
    }
});

// Track filter changes
document.addEventListener('change', function(event) {
    if (event.target.matches('select, input[type="radio"], input[type="checkbox"]')) {
        const filterType = event.target.name || event.target.id;
        const filterValue = event.target.value;
        
        if (filterType && filterValue) {
            analytics.trackEvent('filter_change', 'search_filter', `${filterType}: ${filterValue}`);
        }
    }
});

// Track auth button clicks
document.addEventListener('click', function(event) {
    if (event.target.closest('.login-btn')) {
        analytics.trackAuth('login_attempt');
    }
    if (event.target.closest('.logout')) {
        analytics.trackAuth('logout');
    }
});

// Track video/media interactions
document.addEventListener('play', function(event) {
    if (event.target.tagName === 'VIDEO') {
        const title = event.target.title || event.target.getAttribute('data-title') || 'Unknown Video';
        analytics.trackMediaInteraction('video', 'play', title);
    }
}, true);

document.addEventListener('pause', function(event) {
    if (event.target.tagName === 'VIDEO') {
        const title = event.target.title || event.target.getAttribute('data-title') || 'Unknown Video';
        analytics.trackMediaInteraction('video', 'pause', title);
    }
}, true);

// Make analytics available globally for manual tracking
window.analytics = analytics;
