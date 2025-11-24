/**
 * React Components for Endpoints Module
 *
 * This file contains custom React components for the Lantern Endpoints tab
 * to provide a modern, interactive UI experience.
 */

(function() {
  'use strict';

  // Wait for React to be available
  if (typeof React === 'undefined' || typeof ReactDOM === 'undefined') {
    console.error('React is not loaded. Please ensure React and ReactDOM are included before this script.');
    return;
  }

  const { useState, useEffect, useMemo } = React;

  /**
   * Modern Card Component for displaying endpoint statistics
   */
  const StatCard = ({ title, value, icon, color = '#1B5A7F', loading = false }) => {
    return React.createElement(
      'div',
      {
        className: 'stat-card',
        style: {
          background: 'white',
          borderRadius: '8px',
          padding: '20px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          borderLeft: `4px solid ${color}`,
          minHeight: '120px',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          transition: 'all 0.3s ease',
        }
      },
      [
        React.createElement(
          'div',
          {
            key: 'header',
            style: {
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: '10px'
            }
          },
          [
            React.createElement(
              'h3',
              {
                key: 'title',
                style: {
                  margin: 0,
                  fontSize: '14px',
                  fontWeight: '500',
                  color: '#666',
                  textTransform: 'uppercase',
                  letterSpacing: '0.5px'
                }
              },
              title
            ),
            icon && React.createElement(
              'i',
              {
                key: 'icon',
                className: `fa ${icon}`,
                style: {
                  fontSize: '24px',
                  color: color,
                  opacity: 0.7
                }
              }
            )
          ]
        ),
        React.createElement(
          'div',
          {
            key: 'value',
            style: {
              fontSize: '32px',
              fontWeight: 'bold',
              color: '#333',
              marginTop: '10px'
            }
          },
          loading ? 'Loading...' : value
        )
      ]
    );
  };

  /**
   * Enhanced Search Bar Component with real-time feedback
   */
  const SearchBar = ({ value, onChange, placeholder = 'Search endpoints...', onClear }) => {
    return React.createElement(
      'div',
      {
        style: {
          position: 'relative',
          width: '100%',
          maxWidth: '600px',
          marginBottom: '20px'
        }
      },
      [
        React.createElement('input', {
          key: 'search-input',
          type: 'text',
          value: value,
          onChange: (e) => onChange(e.target.value),
          placeholder: placeholder,
          style: {
            width: '100%',
            padding: '12px 40px 12px 16px',
            fontSize: '16px',
            border: '2px solid #e0e0e0',
            borderRadius: '8px',
            outline: 'none',
            transition: 'border-color 0.3s ease',
            boxSizing: 'border-box'
          },
          onFocus: (e) => {
            e.target.style.borderColor = '#1B5A7F';
          },
          onBlur: (e) => {
            e.target.style.borderColor = '#e0e0e0';
          }
        }),
        React.createElement(
          'i',
          {
            key: 'search-icon',
            className: 'fa fa-search',
            style: {
              position: 'absolute',
              right: '16px',
              top: '50%',
              transform: 'translateY(-50%)',
              color: '#999',
              pointerEvents: 'none'
            }
          }
        ),
        value && React.createElement(
          'button',
          {
            key: 'clear-btn',
            onClick: onClear,
            style: {
              position: 'absolute',
              right: '40px',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: '#999',
              fontSize: '18px',
              padding: '4px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            },
            'aria-label': 'Clear search'
          },
          React.createElement('i', { className: 'fa fa-times' })
        )
      ]
    );
  };

  /**
   * Modern Pagination Component
   */
  const Pagination = ({ currentPage, totalPages, onPageChange, onPrevious, onNext }) => {
    const [inputPage, setInputPage] = useState(currentPage);

    useEffect(() => {
      setInputPage(currentPage);
    }, [currentPage]);

    const handleInputChange = (e) => {
      const value = e.target.value;
      setInputPage(value);
    };

    const handleInputBlur = () => {
      const pageNum = parseInt(inputPage);
      if (!isNaN(pageNum) && pageNum >= 1 && pageNum <= totalPages) {
        onPageChange(pageNum);
      } else {
        setInputPage(currentPage);
      }
    };

    const handleKeyPress = (e) => {
      if (e.key === 'Enter') {
        handleInputBlur();
      }
    };

    return React.createElement(
      'div',
      {
        style: {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginTop: '20px',
          padding: '20px 0',
          borderTop: '1px solid #e0e0e0'
        }
      },
      [
        React.createElement(
          'button',
          {
            key: 'prev-btn',
            onClick: onPrevious,
            disabled: currentPage <= 1,
            className: 'btn btn-default',
            style: {
              padding: '8px 16px',
              borderRadius: '6px',
              border: '1px solid #ddd',
              background: currentPage <= 1 ? '#f5f5f5' : 'white',
              cursor: currentPage <= 1 ? 'not-allowed' : 'pointer',
              opacity: currentPage <= 1 ? 0.5 : 1,
              transition: 'all 0.3s ease'
            }
          },
          [
            React.createElement('i', { key: 'icon', className: 'fa fa-arrow-left', style: { marginRight: '8px' } }),
            React.createElement('span', { key: 'text' }, 'Previous')
          ]
        ),
        React.createElement(
          'div',
          {
            key: 'page-info',
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
              fontSize: '16px'
            }
          },
          [
            React.createElement('span', { key: 'page-label', style: { color: '#666' } }, 'Page'),
            React.createElement('input', {
              key: 'page-input',
              type: 'number',
              value: inputPage,
              onChange: handleInputChange,
              onBlur: handleInputBlur,
              onKeyPress: handleKeyPress,
              min: 1,
              max: totalPages,
              style: {
                width: '60px',
                padding: '6px 8px',
                textAlign: 'center',
                border: '1px solid #ddd',
                borderRadius: '4px',
                fontSize: '16px'
              }
            }),
            React.createElement('span', { key: 'total-pages', style: { color: '#666' } }, `of ${totalPages}`)
          ]
        ),
        React.createElement(
          'button',
          {
            key: 'next-btn',
            onClick: onNext,
            disabled: currentPage >= totalPages,
            className: 'btn btn-default',
            style: {
              padding: '8px 16px',
              borderRadius: '6px',
              border: '1px solid #ddd',
              background: currentPage >= totalPages ? '#f5f5f5' : 'white',
              cursor: currentPage >= totalPages ? 'not-allowed' : 'pointer',
              opacity: currentPage >= totalPages ? 0.5 : 1,
              transition: 'all 0.3s ease'
            }
          },
          [
            React.createElement('span', { key: 'text', style: { marginRight: '8px' } }, 'Next'),
            React.createElement('i', { key: 'icon', className: 'fa fa-arrow-right' })
          ]
        )
      ]
    );
  };

  /**
   * Action Buttons Component
   */
  const ActionButtons = ({ onDownloadData, onDownloadDescriptions, loading = false }) => {
    return React.createElement(
      'div',
      {
        style: {
          display: 'flex',
          gap: '12px',
          flexWrap: 'wrap',
          marginBottom: '20px'
        }
      },
      [
        React.createElement(
          'button',
          {
            key: 'download-data',
            onClick: onDownloadData,
            disabled: loading,
            className: 'btn btn-primary',
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '10px 20px',
              borderRadius: '6px',
              border: 'none',
              background: '#1B5A7F',
              color: 'white',
              fontSize: '14px',
              fontWeight: '500',
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.7 : 1,
              transition: 'all 0.3s ease'
            },
            onMouseEnter: (e) => {
              if (!loading) e.target.style.background = '#164a68';
            },
            onMouseLeave: (e) => {
              e.target.style.background = '#1B5A7F';
            }
          },
          [
            React.createElement('i', { key: 'icon', className: 'fa fa-download' }),
            React.createElement('span', { key: 'text' }, 'Download Endpoint Data (CSV)')
          ]
        ),
        React.createElement(
          'button',
          {
            key: 'download-descriptions',
            onClick: onDownloadDescriptions,
            disabled: loading,
            className: 'btn btn-default',
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '10px 20px',
              borderRadius: '6px',
              border: '1px solid #ddd',
              background: 'white',
              color: '#333',
              fontSize: '14px',
              fontWeight: '500',
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.7 : 1,
              transition: 'all 0.3s ease'
            },
            onMouseEnter: (e) => {
              if (!loading) e.target.style.background = '#f5f5f5';
            },
            onMouseLeave: (e) => {
              e.target.style.background = 'white';
            }
          },
          [
            React.createElement('i', { key: 'icon', className: 'fa fa-file-text' }),
            React.createElement('span', { key: 'text' }, 'Download Field Descriptions (CSV)')
          ]
        )
      ]
    );
  };

  /**
   * Info Banner Component
   */
  const InfoBanner = ({ type = 'info', message, icon = 'fa-info-circle' }) => {
    const colors = {
      info: { bg: '#e3f2fd', border: '#2196f3', text: '#1565c0' },
      warning: { bg: '#fff3e0', border: '#ff9800', text: '#e65100' },
      success: { bg: '#e8f5e9', border: '#4caf50', text: '#2e7d32' }
    };

    const color = colors[type] || colors.info;

    return React.createElement(
      'div',
      {
        style: {
          display: 'flex',
          alignItems: 'flex-start',
          gap: '12px',
          padding: '16px',
          borderRadius: '8px',
          background: color.bg,
          border: `1px solid ${color.border}`,
          marginBottom: '20px'
        },
        role: 'alert'
      },
      [
        React.createElement('i', {
          key: 'icon',
          className: `fa ${icon}`,
          style: {
            fontSize: '20px',
            color: color.text,
            marginTop: '2px'
          }
        }),
        React.createElement(
          'div',
          {
            key: 'message',
            style: {
              flex: 1,
              color: color.text,
              fontSize: '14px',
              lineHeight: '1.6'
            }
          },
          message
        )
      ]
    );
  };

  /**
   * Loading Skeleton Component
   */
  const LoadingSkeleton = ({ rows = 5 }) => {
    const skeletonRows = Array.from({ length: rows }, (_, i) => i);

    return React.createElement(
      'div',
      {
        style: {
          padding: '20px',
          background: 'white',
          borderRadius: '8px'
        }
      },
      skeletonRows.map((_, index) =>
        React.createElement(
          'div',
          {
            key: index,
            style: {
              height: '40px',
              background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
              backgroundSize: '200% 100%',
              animation: 'loading 1.5s infinite',
              borderRadius: '4px',
              marginBottom: '10px'
            }
          }
        )
      )
    );
  };

  // Export components to global scope for use in R
  window.LanternReactComponents = {
    StatCard,
    SearchBar,
    Pagination,
    ActionButtons,
    InfoBanner,
    LoadingSkeleton
  };

  // Add CSS animation for loading skeleton
  if (!document.getElementById('lantern-react-animations')) {
    const style = document.createElement('style');
    style.id = 'lantern-react-animations';
    style.textContent = `
      @keyframes loading {
        0% { background-position: 200% 0; }
        100% { background-position: -200% 0; }
      }

      .stat-card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0,0,0,0.15) !important;
      }
    `;
    document.head.appendChild(style);
  }

  console.log('Lantern React Components loaded successfully');
})();
