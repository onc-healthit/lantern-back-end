/**
 * React Components for Security Module using shiny.react
 *
 * This file contains custom React components for the Lantern Security tab
 * designed to work with the shiny.react package.
 */

(function() {
  'use strict';

  // Wait for React to be available
  if (typeof React === 'undefined' || typeof ReactDOM === 'undefined') {
    console.error('React is not loaded. Please ensure React and ReactDOM are included before this script.');
    return;
  }

  const { useState, useEffect, useMemo, useCallback } = React;

  /**
   * Summary Card Component for displaying security statistics
   */
  const SecuritySummaryCard = ({ title, count, percentage, icon, color = '#1B5A7F' }) => {
    return React.createElement(
      'div',
      {
        className: 'security-summary-card',
        style: {
          background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
          borderRadius: '12px',
          padding: '24px',
          boxShadow: '0 4px 16px rgba(0,0,0,0.1)',
          borderLeft: `5px solid ${color}`,
          minHeight: '140px',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          transition: 'all 0.3s ease',
          position: 'relative',
          overflow: 'hidden'
        }
      },
      [
        // Background icon
        React.createElement('i', {
          key: 'bg-icon',
          className: `fa ${icon}`,
          style: {
            position: 'absolute',
            right: '-20px',
            top: '-20px',
            fontSize: '120px',
            color: color,
            opacity: 0.05,
            transform: 'rotate(-15deg)'
          }
        }),
        // Header
        React.createElement(
          'div',
          {
            key: 'header',
            style: {
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: '16px',
              position: 'relative',
              zIndex: 1
            }
          },
          [
            React.createElement(
              'h3',
              {
                key: 'title',
                style: {
                  margin: 0,
                  fontSize: '15px',
                  fontWeight: '600',
                  color: '#555',
                  textTransform: 'uppercase',
                  letterSpacing: '0.5px'
                }
              },
              title
            ),
            icon && React.createElement(
              'div',
              {
                key: 'icon-container',
                style: {
                  width: '48px',
                  height: '48px',
                  borderRadius: '50%',
                  background: `${color}15`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }
              },
              React.createElement('i', {
                className: `fa ${icon}`,
                style: {
                  fontSize: '24px',
                  color: color
                }
              })
            )
          ]
        ),
        // Content
        React.createElement(
          'div',
          {
            key: 'content',
            style: {
              position: 'relative',
              zIndex: 1
            }
          },
          [
            React.createElement(
              'div',
              {
                key: 'count',
                style: {
                  fontSize: '36px',
                  fontWeight: 'bold',
                  color: '#333',
                  marginBottom: '8px'
                }
              },
              count !== undefined ? count.toLocaleString() : '-'
            ),
            percentage !== undefined && React.createElement(
              'div',
              {
                key: 'percentage',
                style: {
                  fontSize: '14px',
                  color: '#666',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px'
                }
              },
              [
                React.createElement(
                  'span',
                  {
                    key: 'badge',
                    style: {
                      background: `${color}20`,
                      color: color,
                      padding: '2px 8px',
                      borderRadius: '4px',
                      fontSize: '12px',
                      fontWeight: '600'
                    }
                  },
                  `${percentage}%`
                ),
                React.createElement('span', { key: 'label' }, 'of total')
              ]
            )
          ]
        )
      ]
    );
  };

  /**
   * Auth Type Badge Component
   */
  const AuthTypeBadge = ({ type, count, isActive, onClick }) => {
    const colors = {
      'SMART-on-FHIR': { bg: '#4CAF50', text: '#ffffff' },
      'OAuth': { bg: '#2196F3', text: '#ffffff' },
      'Basic': { bg: '#FF9800', text: '#ffffff' },
      'Bearer': { bg: '#9C27B0', text: '#ffffff' },
      'None': { bg: '#757575', text: '#ffffff' }
    };

    const color = colors[type] || { bg: '#1B5A7F', text: '#ffffff' };

    return React.createElement(
      'button',
      {
        onClick: onClick,
        style: {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          width: '100%',
          padding: '14px 18px',
          marginBottom: '10px',
          background: isActive ? color.bg : '#ffffff',
          color: isActive ? color.text : '#333',
          border: isActive ? `2px solid ${color.bg}` : '2px solid #e0e0e0',
          borderRadius: '8px',
          cursor: 'pointer',
          fontSize: '15px',
          fontWeight: isActive ? '600' : '500',
          transition: 'all 0.3s ease',
          boxShadow: isActive ? '0 4px 12px rgba(0,0,0,0.15)' : '0 2px 4px rgba(0,0,0,0.05)',
          transform: isActive ? 'translateY(-2px)' : 'none'
        },
        onMouseEnter: (e) => {
          if (!isActive) {
            e.target.style.borderColor = color.bg;
            e.target.style.boxShadow = '0 4px 8px rgba(0,0,0,0.1)';
          }
        },
        onMouseLeave: (e) => {
          if (!isActive) {
            e.target.style.borderColor = '#e0e0e0';
            e.target.style.boxShadow = '0 2px 4px rgba(0,0,0,0.05)';
          }
        }
      },
      [
        React.createElement(
          'span',
          {
            key: 'type',
            style: {
              display: 'flex',
              alignItems: 'center',
              gap: '10px'
            }
          },
          [
            React.createElement('i', {
              key: 'icon',
              className: 'fa fa-shield',
              style: {
                fontSize: '18px'
              }
            }),
            React.createElement('span', { key: 'text' }, type)
          ]
        ),
        React.createElement(
          'span',
          {
            key: 'count',
            style: {
              background: isActive ? 'rgba(255,255,255,0.3)' : color.bg + '20',
              color: isActive ? color.text : color.bg,
              padding: '4px 12px',
              borderRadius: '12px',
              fontSize: '14px',
              fontWeight: 'bold',
              minWidth: '40px',
              textAlign: 'center'
            }
          },
          count
        )
      ]
    );
  };

  /**
   * Modern Data Table Component
   */
  const SecurityDataTable = ({ data, columns, loading = false }) => {
    const [sortColumn, setSortColumn] = useState(null);
    const [sortDirection, setSortDirection] = useState('asc');

    const handleSort = useCallback((columnKey) => {
      if (sortColumn === columnKey) {
        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
      } else {
        setSortColumn(columnKey);
        setSortDirection('asc');
      }
    }, [sortColumn, sortDirection]);

    const sortedData = useMemo(() => {
      if (!sortColumn || !data) return data;

      return [...data].sort((a, b) => {
        const aVal = a[sortColumn];
        const bVal = b[sortColumn];

        if (aVal === bVal) return 0;

        const comparison = aVal < bVal ? -1 : 1;
        return sortDirection === 'asc' ? comparison : -comparison;
      });
    }, [data, sortColumn, sortDirection]);

    if (loading) {
      return React.createElement(
        'div',
        {
          style: {
            padding: '40px',
            textAlign: 'center',
            background: '#f8f9fa',
            borderRadius: '8px'
          }
        },
        [
          React.createElement('i', {
            key: 'icon',
            className: 'fa fa-spinner fa-spin',
            style: { fontSize: '32px', color: '#1B5A7F', marginBottom: '16px' }
          }),
          React.createElement(
            'div',
            {
              key: 'text',
              style: { fontSize: '16px', color: '#666' }
            },
            'Loading security data...'
          )
        ]
      );
    }

    if (!data || data.length === 0) {
      return React.createElement(
        'div',
        {
          style: {
            padding: '40px',
            textAlign: 'center',
            background: '#f8f9fa',
            borderRadius: '8px'
          }
        },
        [
          React.createElement('i', {
            key: 'icon',
            className: 'fa fa-inbox',
            style: { fontSize: '48px', color: '#ccc', marginBottom: '16px' }
          }),
          React.createElement(
            'div',
            {
              key: 'text',
              style: { fontSize: '16px', color: '#999' }
            },
            'No data available'
          )
        ]
      );
    }

    return React.createElement(
      'div',
      {
        style: {
          background: 'white',
          borderRadius: '12px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
          overflow: 'hidden'
        }
      },
      [
        // Table
        React.createElement(
          'div',
          {
            key: 'table-container',
            style: {
              overflowX: 'auto'
            }
          },
          React.createElement(
            'table',
            {
              style: {
                width: '100%',
                borderCollapse: 'collapse',
                fontSize: '14px'
              }
            },
            [
              // Header
              React.createElement(
                'thead',
                {
                  key: 'thead',
                  style: {
                    background: '#f6f7f8',
                    borderBottom: '2px solid #1B5A7F'
                  }
                },
                React.createElement(
                  'tr',
                  null,
                  columns.map((col, idx) =>
                    React.createElement(
                      'th',
                      {
                        key: idx,
                        onClick: () => handleSort(col.key),
                        style: {
                          padding: '16px 12px',
                          textAlign: col.align || 'left',
                          fontWeight: '600',
                          color: '#333',
                          cursor: col.sortable !== false ? 'pointer' : 'default',
                          whiteSpace: 'nowrap',
                          userSelect: 'none'
                        }
                      },
                      [
                        React.createElement('span', { key: 'label' }, col.label),
                        col.sortable !== false && sortColumn === col.key && React.createElement(
                          'i',
                          {
                            key: 'icon',
                            className: `fa fa-sort-${sortDirection === 'asc' ? 'up' : 'down'}`,
                            style: { marginLeft: '8px', color: '#1B5A7F' }
                          }
                        )
                      ]
                    )
                  )
                )
              ),
              // Body
              React.createElement(
                'tbody',
                {
                  key: 'tbody'
                },
                sortedData.map((row, rowIdx) =>
                  React.createElement(
                    'tr',
                    {
                      key: rowIdx,
                      style: {
                        borderBottom: '1px solid #e0e0e0',
                        background: rowIdx % 2 === 0 ? '#ffffff' : '#f9f9f9',
                        transition: 'background 0.2s ease'
                      },
                      onMouseEnter: (e) => {
                        e.currentTarget.style.background = '#f0f7ff';
                      },
                      onMouseLeave: (e) => {
                        e.currentTarget.style.background = rowIdx % 2 === 0 ? '#ffffff' : '#f9f9f9';
                      }
                    },
                    columns.map((col, colIdx) =>
                      React.createElement(
                        'td',
                        {
                          key: colIdx,
                          style: {
                            padding: '14px 12px',
                            textAlign: col.align || 'left',
                            color: '#555'
                          },
                          dangerouslySetInnerHTML: col.html ? { __html: row[col.key] } : undefined
                        },
                        !col.html && (col.render ? col.render(row[col.key], row) : row[col.key])
                      )
                    )
                  )
                )
              )
            ]
          )
        )
      ]
    );
  };

  /**
   * Enhanced Search Component with filters
   */
  const SecuritySearchBar = ({ value, onChange, placeholder = 'Search security endpoints...', onClear }) => {
    return React.createElement(
      'div',
      {
        style: {
          position: 'relative',
          width: '100%',
          marginBottom: '24px'
        }
      },
      [
        React.createElement('input', {
          key: 'input',
          type: 'text',
          value: value,
          onChange: (e) => onChange(e.target.value),
          placeholder: placeholder,
          style: {
            width: '100%',
            padding: '14px 50px 14px 46px',
            fontSize: '16px',
            border: '2px solid #e0e0e0',
            borderRadius: '10px',
            outline: 'none',
            transition: 'all 0.3s ease',
            boxSizing: 'border-box',
            background: '#ffffff'
          },
          onFocus: (e) => {
            e.target.style.borderColor = '#1B5A7F';
            e.target.style.boxShadow = '0 4px 12px rgba(27, 90, 127, 0.1)';
          },
          onBlur: (e) => {
            e.target.style.borderColor = '#e0e0e0';
            e.target.style.boxShadow = 'none';
          }
        }),
        React.createElement(
          'i',
          {
            key: 'search-icon',
            className: 'fa fa-search',
            style: {
              position: 'absolute',
              left: '18px',
              top: '50%',
              transform: 'translateY(-50%)',
              color: '#999',
              fontSize: '18px',
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
              right: '14px',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: '#999',
              fontSize: '20px',
              padding: '6px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: '50%',
              transition: 'all 0.2s ease'
            },
            onMouseEnter: (e) => {
              e.target.style.background = '#f0f0f0';
              e.target.style.color = '#333';
            },
            onMouseLeave: (e) => {
              e.target.style.background = 'none';
              e.target.style.color = '#999';
            },
            'aria-label': 'Clear search'
          },
          React.createElement('i', { className: 'fa fa-times-circle' })
        )
      ]
    );
  };

  // Export components to global scope
  window.SecurityReactComponents = {
    SecuritySummaryCard,
    AuthTypeBadge,
    SecurityDataTable,
    SecuritySearchBar
  };

  // Add CSS animations
  if (!document.getElementById('security-react-animations')) {
    const style = document.createElement('style');
    style.id = 'security-react-animations';
    style.textContent = `
      @keyframes fadeIn {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
      }

      .security-summary-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 24px rgba(0,0,0,0.15) !important;
      }

      @media (max-width: 768px) {
        .security-summary-card {
          margin-bottom: 16px;
        }
      }
    `;
    document.head.appendChild(style);
  }

  console.log('Security React Components loaded successfully');
})();
