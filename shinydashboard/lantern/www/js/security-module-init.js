/**
 * Security Module Initialization Script
 * Activates all React components for the Security tab
 */

(function() {
  'use strict';

  window.initializeSecurityReactComponents = function(ns) {
    // Check if React components are loaded
    if (typeof window.SecurityReactComponents === 'undefined' ||
        typeof React === 'undefined' ||
        typeof ReactDOM === 'undefined') {
      console.error('React or SecurityReactComponents not loaded');
      return;
    }

    const { SecuritySummaryCard, SecuritySearchBar, AuthTypeBadge } = window.SecurityReactComponents;

    // ============================================
    // 1. RENDER SUMMARY CARDS
    // ============================================
    const summaryContainer = document.getElementById(ns + 'react_summary_cards_container');
    console.log('[Security Init] Summary container element ID:', ns + 'react_summary_cards_container');
    console.log('[Security Init] Summary container found:', !!summaryContainer);

    // Initialize only if not already initialized
    if (summaryContainer && !window._securitySummaryCardRoots) {
      console.log('[Security Init] Initializing summary cards for namespace:', ns);

      // Clear placeholder content
      summaryContainer.innerHTML = '';

      // Create a grid for cards
      const cardsGrid = document.createElement('div');
      cardsGrid.id = ns + 'cards_grid';
      cardsGrid.style.cssText = 'display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; margin-bottom: 24px;';

      // Card configuration
      const summaryCards = [
        {
          title: 'Total Endpoints',
          count: 0,
          icon: 'fa-database',
          color: '#1B5A7F',
          id: 'total_endpoints_card'
        },
        {
          title: 'With Security',
          count: 0,
          icon: 'fa-shield',
          color: '#4CAF50',
          id: 'with_security_card'
        },
        {
          title: 'Without Security',
          count: 0,
          icon: 'fa-unlock',
          color: '#FF9800',
          id: 'without_security_card'
        },
        {
          title: 'Auth Types',
          count: 0,
          icon: 'fa-key',
          color: '#9C27B0',
          id: 'auth_types_card'
        }
      ];

      // Store React roots globally (critical for React 18 - must reuse roots)
      window._securitySummaryCardRoots = {};
      window._securitySummaryCardsConfig = summaryCards;
      window._securitySummaryCardsNS = ns;

      summaryCards.forEach(card => {
        const cardDiv = document.createElement('div');
        cardDiv.id = ns + card.id;

        // Render using React 18 API
        const root = ReactDOM.createRoot(cardDiv);
        window._securitySummaryCardRoots[card.id] = root;

        // Initial render
        root.render(
          React.createElement(SecuritySummaryCard, {
            title: card.title,
            count: card.count,
            icon: card.icon,
            color: card.color
          })
        );

        cardsGrid.appendChild(cardDiv);
      });

      summaryContainer.appendChild(cardsGrid);
      console.log('[Security Init] Summary cards rendered successfully. Card count:', summaryCards.length);
    } else if (summaryContainer && window._securitySummaryCardRoots) {
      console.log('[Security Init] Summary cards already initialized');
    } else if (!summaryContainer) {
      console.error('[Security Init] Summary cards container not found! Looking for:', ns + 'react_summary_cards_container');
      console.error('[Security Init] Available elements with "react" in ID:',
        Array.from(document.querySelectorAll('[id*="react"]')).map(el => el.id));
    }

    // ============================================
    // 2. RENDER AUTH TYPE BADGES
    // ============================================
    const badgesContainer = document.getElementById(ns + 'auth_type_badges_container');
    if (badgesContainer && !badgesContainer.hasChildNodes()) {
      const badgesGrid = document.createElement('div');
      badgesGrid.style.cssText = 'display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; margin-bottom: 20px;';
      badgesGrid.id = ns + 'auth_badges_grid';

      badgesContainer.appendChild(badgesGrid);

      // Store for later updates
      window._securityAuthBadgesNS = ns;
    }

    // ============================================
    // 3. RENDER ENHANCED SEARCH BAR
    // ============================================
    const searchContainer = document.getElementById(ns + 'security_search_container');
    const searchInput = document.getElementById(ns + 'security_search_query');

    if (searchContainer && searchInput) {
      // Create a new div for React search bar
      const reactSearchDiv = document.createElement('div');
      reactSearchDiv.id = ns + 'react_search_bar';
      reactSearchDiv.style.marginBottom = '24px';

      const root = ReactDOM.createRoot(reactSearchDiv);

      // Initial render
      const updateSearch = (value) => {
        // Update the hidden Shiny input
        searchInput.value = value;
        searchInput.dispatchEvent(new Event('input', { bubbles: true }));
      };

      const clearSearch = () => {
        updateSearch('');
      };

      const renderSearch = (value) => {
        root.render(
          React.createElement(SecuritySearchBar, {
            value: value || '',
            onChange: updateSearch,
            placeholder: 'Search by URL, organization, vendor, FHIR version, or TLS version...',
            onClear: clearSearch
          })
        );
      };

      renderSearch(searchInput.value);

      // Listen for external changes to Shiny input and update React
      searchInput.addEventListener('input', function() {
        renderSearch(this.value);
      });

      searchContainer.appendChild(reactSearchDiv);
    }

    // ============================================
    // 4. UPDATE CARDS WITH REAL DATA FROM SHINY
    // ============================================

    // Listen for Shiny to send updated summary data
    if (typeof Shiny !== 'undefined') {
      const handlerName1 = ns + 'update_summary_cards';
      const handlerName2 = ns + 'update_auth_badges';

      console.log('Registering Shiny message handlers:', handlerName1, handlerName2);

      // Remove old handler if it exists
      if (Shiny.shinyapp.config.customMessageHandlers[handlerName1]) {
        delete Shiny.shinyapp.config.customMessageHandlers[handlerName1];
      }

      Shiny.addCustomMessageHandler(handlerName1, function(data) {
        console.log('Received summary card data:', data);

        const cards = {
          'total_endpoints_card': data.total || 0,
          'with_security_card': data.withSecurity || 0,
          'without_security_card': data.withoutSecurity || 0,
          'auth_types_card': data.authTypes || 0
        };

        const summaryCards = window._securitySummaryCardsConfig || [];
        const cardRoots = window._securitySummaryCardRoots || {};

        console.log('Available card roots:', Object.keys(cardRoots));
        console.log('Card data to update:', cards);

        Object.keys(cards).forEach(cardId => {
          const cardConfig = summaryCards.find(c => c.id === cardId);
          const root = cardRoots[cardId];

          if (root && cardConfig) {
            // REUSE the existing root for updates (React 18 requirement)
            root.render(
              React.createElement(SecuritySummaryCard, {
                title: cardConfig.title,
                count: cards[cardId],
                icon: cardConfig.icon,
                color: cardConfig.color
              })
            );
            console.log('Updated card:', cardId, 'with count:', cards[cardId]);
          } else {
            console.error('Missing root or config for card:', cardId, 'root exists:', !!root, 'config exists:', !!cardConfig);
          }
        });
      });

      // Remove old handler if it exists
      if (Shiny.shinyapp.config.customMessageHandlers[handlerName2]) {
        delete Shiny.shinyapp.config.customMessageHandlers[handlerName2];
      }

      // Listen for auth type data
      Shiny.addCustomMessageHandler(handlerName2, function(authData) {
        console.log('Received auth badges data:', authData);

        // Store auth data for later updates
        window._securityLastAuthData = authData;

        const badgeNS = window._securityAuthBadgesNS || ns;
        const badgesGrid = document.getElementById(badgeNS + 'auth_badges_grid');

        if (badgesGrid) {
          // Clear existing badges and roots
          badgesGrid.innerHTML = '';

          // Store roots for auth badges too
          if (!window._securityAuthBadgeRoots) {
            window._securityAuthBadgeRoots = [];
          }

          authData.forEach((auth, index) => {
            const badgeDiv = document.createElement('div');
            badgeDiv.id = badgeNS + 'auth_badge_' + index;
            const root = ReactDOM.createRoot(badgeDiv);

            root.render(
              React.createElement(AuthTypeBadge, {
                type: auth.type,
                count: auth.count,
                isActive: auth.isActive || false,
                onClick: () => {
                  console.log('Auth type clicked:', auth.type);

                  // Update the Shiny dropdown select input
                  const authDropdown = document.getElementById('auth_type_code');
                  if (authDropdown) {
                    authDropdown.value = auth.type;
                    authDropdown.dispatchEvent(new Event('change', { bubbles: true }));
                  }
                }
              })
            );
            badgesGrid.appendChild(badgeDiv);
          });

          console.log('Auth badges rendered:', authData.length, 'badges');
        } else {
          console.error('Auth badges grid not found:', badgeNS + 'auth_badges_grid');
        }
      });

      // Listen for changes to auth_type_code dropdown and update badges
      // Use a polling mechanism since the dropdown is rendered by Shiny dynamically
      const setupDropdownListener = () => {
        const authDropdown = document.getElementById('auth_type_code');
        if (authDropdown && !authDropdown._securityListenerAttached) {
          authDropdown._securityListenerAttached = true;

          authDropdown.addEventListener('change', function() {
            const selectedAuthType = this.value;

            // Update active state of all badges
            const badgesGrid = document.getElementById(ns + 'auth_badges_grid');
            if (badgesGrid && window._securityLastAuthData) {
              badgesGrid.innerHTML = '';

              window._securityLastAuthData.forEach((auth, index) => {
                const badgeDiv = document.createElement('div');
                badgeDiv.id = ns + 'auth_badge_' + index;
                const root = ReactDOM.createRoot(badgeDiv);
                root.render(
                  React.createElement(AuthTypeBadge, {
                    type: auth.type,
                    count: auth.count,
                    isActive: auth.type === selectedAuthType,
                    onClick: () => {
                      authDropdown.value = auth.type;
                      authDropdown.dispatchEvent(new Event('change', { bubbles: true }));
                    }
                  })
                );
                badgesGrid.appendChild(badgeDiv);
              });
            }
          });

          console.log('Auth type dropdown listener attached');
        } else if (!authDropdown) {
          // Retry after a short delay
          setTimeout(setupDropdownListener, 500);
        }
      };

      // Start trying to attach the listener
      setupDropdownListener();
    }

    console.log('Security React components initialized for namespace:', ns);
  };
})();
