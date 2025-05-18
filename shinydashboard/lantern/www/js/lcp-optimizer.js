// LCP optimizer script for Lantern Dashboard
(function() {
    // Define the most likely LCP elements
    var lcpSelectors = ['h2', '#dashboard_page h2', '#prerendered-h2'];
    
    // Function to apply optimizations to potential LCP elements
    function optimizeLCP() {
      for (var i = 0; i < lcpSelectors.length; i++) {
        var elements = document.querySelectorAll(lcpSelectors[i]);
        
        for (var j = 0; j < elements.length; j++) {
          var el = elements[j];
          // Apply high-priority rendering optimizations
          if (el && el.style) {
            el.style.visibility = 'visible';
            el.style.display = 'block';
            
            // Apply advanced rendering optimizations for modern browsers
            if ('contentVisibility' in el.style) {
              el.style.contentVisibility = 'auto';
              el.style.contain = 'content';
            }
          }
        }
      }
    }
    
    // Run immediately
    optimizeLCP();
    
    // Also run when DOM is ready
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', optimizeLCP);
    }
  
    // Monitor for LCP
    if (window.PerformanceObserver) {
      try {
        var lcpObserver = new PerformanceObserver(function(list) {
          var entries = list.getEntries();
          if (entries && entries.length > 0) {
            var lastEntry = entries[entries.length - 1];
            if (lastEntry && lastEntry.element && lastEntry.element.style) {
              lastEntry.element.style.visibility = 'visible';
              lastEntry.element.style.display = 'block';
            }
          }
        });
        
        lcpObserver.observe({type: 'largest-contentful-paint', buffered: true});
      } catch (e) {
        // Silent fail
      }
    }
  })();