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
    // Smooth scrolling for CTA button
    const ctaButton = document.querySelector('.cta-button');
    if (ctaButton) {
        ctaButton.addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('featured').scrollIntoView({
                behavior: 'smooth'
            });
        });
    }
    
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
