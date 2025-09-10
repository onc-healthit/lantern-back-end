// Add hidden skip to content button at the top of the Lantern website
$(document).ready(function() {
    $("header").find("nav").prepend("<a href='#content' aria-label='Click the enter key to skip to the main content of this page, skipping over the header elements and navigation tabs.' class='show-on-focus'>Skip to Content</a>");
  });
  
  // Set the role of Lantern's main content containing element to 'main'
  let elems = document.getElementsByClassName('content-wrapper');
  elems[0].setAttribute('role', 'main');
  elems[0].id = 'content';
  
  document.getElementById('side_menu').classList.add('sidebarMenuSelectedTabItem');
  
  /*
    // Create an observer that watches if an element's tabindex attribute has been set or altered
    // If the altered attribute is the tabindex attribute, and its value is not -5, remove the tabindex attribute
    // Tabindex is removed from elements for two different reasons: 
      // 1. R Shiny automatically adds a tabindex of some elements, like the sidebar tab items, to -1 when they are clicked off of, meaning you can no longer tab to those items. 
            Removing the tabindex solves this problem. Some elements are focusable by default, and thus we can remove the tab index attribute entirely.
      // 2. R Shiny automatically adds some tabindex attributes to elements we do not want to be focusable, so the attribute is removed on these elements.
  */
  let tabIndexObserver = new MutationObserver(function(mutations) {
    for (let mutation of mutations) {
      if (mutation.type === "attributes") {
        if (mutation.target.hasAttribute("tabindex") && mutation.target.getAttribute("tabindex") !== "-5") {
          mutation.target.removeAttribute("tabindex");
        }
      }
    }
  });
  
  // Set the tabindex observer on each of Lantern's tab pages which will remove the tab index attribute whenever it is set or altered 
  let tabPanes = document.getElementsByClassName("tab-pane");
  for (let tab of tabPanes) {
    tabIndexObserver.observe(tab, {
      attributes: true,
      attributeFilter: ["tabindex"]
    });
  }
  
  let sideMenu = document.getElementsByClassName("sidebar-menu");
  let sideMenuList = sideMenu[0].getElementsByTagName("li");
  
  /*
    // Set the tabindex observer on each of Lantern's sidebar tabs which will remove the tab index attribute
    // whenever it is set due to clicking on another tab
  */ 
  for (let liElem of sideMenuList) {
    let sideMenuLinks = liElem.getElementsByTagName("a");
    if (sideMenuLinks.length > 0) {
      for (let aElem of sideMenuLinks) {
        tabIndexObserver.observe(aElem, {
          attributes: true,
          attributeFilter: ["tabindex"]
        });
      }
    }
  }
  
  /*
    // Create an observer that watches for attribute changes on the sidebar to catch when it is opened or closed
    // If the sidebar is closed, set the tab index to -5 so that it is not tabbable
    // If the sidebar is open, remove the tab index attribute
  */
  let sideBarCollapsedObserver = new MutationObserver(function(mutations) {
    for (let mutation of mutations) {
      if (mutation.type === "attributes") {
        let sidebarMenu = document.getElementsByClassName("sidebar-menu");
        let sidebarMenuList = sidebarMenu[0].getElementsByTagName("li");
        for (let liElem of sidebarMenuList) {
          let sidebarMenuLinks = liElem.getElementsByTagName("a");
          if (sidebarMenuLinks.length > 0) {
            for (let aElem of sidebarMenuLinks) {
              if (mutation.target.getAttribute("data-collapsed") === "true") {
                aElem.setAttribute("tabindex", "-5");
              } else {
                aElem.removeAttribute("tabindex");
              }
            }
          }
        }
      }
    }
  });
  
  /* 
    // Set the tabindex observer on the sidebar element which will remove the tabindex attribute
    // from the sidebar elements when the sidebar is opened, and it will add a tabindex of -5 when the sidebar 
    // is collapsed
  */
  let sideBarCollapsed = document.getElementById("sidebarCollapsed");
  sideBarCollapsedObserver.observe(sideBarCollapsed, {
    attributes: true,
    attributeFilter: ["data-collapsed"]
  });
  
  // Create an observer that watches for when any new nodes are created or any existing elements are altered
  let newNodesObserver = new MutationObserver(function(mutations) {   
    for (let mutation of mutations) {
      if (mutation.addedNodes.length > 0) {
        for (let newNode of mutation.addedNodes) {
          
          /*
            // If the mutated node is the response time date filter within the endpoint modal popup and a new node was added to it which was a row element, this means the row containing the date filter dropdown was added
            // Adds a new aria label to the date filter dropdown that explains you can use the arrow keys to navigate the filter dropdown
          */
          if (mutation.target.id === "show_date_filters" && newNode.classList && newNode.classList.contains("row")) {
            let selectInputNodes = newNode.querySelectorAll("select.shiny-bound-input");
            for (let selectInputNode of selectInputNodes) {
              selectInputNode.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.');
            }
          }
  
          /*
            // If the mutated node is the http response date filter within the endpoint modal popup and a new node was added to it which was a row element, this means the row containing the http date filter dropdown was added
            // Adds a new aria label to the date filter dropdown that explains you can use the arrow keys to navigate the filter dropdown
          */
          if (mutation.target.id === "show_http_date_filters" && newNode.classList && newNode.classList.contains("row")) {
            let selectInputNodes = newNode.querySelectorAll("select.shiny-bound-input");
            for (let selectInputNode of selectInputNodes) {
              selectInputNode.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.');
            }
          }
          
          // Checks if the new node that was added was the shiny modal popup 
          if (newNode.id === "shiny-modal-wrapper") {
            
            // Set the tabindex observer on each of the modal popup's tab pages which will remove the tab index attribute whenever it is set or altered 
            let modalTabPanes = newNode.getElementsByClassName("tab-pane");
            for (let tab of modalTabPanes) {
              tabIndexObserver.observe(tab, {
                attributes: true,
                attributeFilter: ["tabindex"]
              });
            }
            
            /* 
              // Get all the navigation bar tabs on the endpoint modal popup and alter them to have the same 
              // structure, classes, and attributes as the navigation bars on the rest of the Lantern pages
            */
            let navBarTabs = document.getElementsByClassName("nav nav-tabs");
            for (let navTab of navBarTabs) {
              
              // Add the tablist role to the main navbar containing element
              navTab.setAttribute("role", "tablist");
              
              let navTabID = navTab.getAttribute("data-tabsetid");
              
              // Add the shiny-tab-input and shiny-bound-input classes to the main navbar containing element
              navTab.classList.add("shiny-tab-input", "shiny-bound-input");
              
              let liElements = navTab.getElementsByTagName("li");
              for (let liElem of liElements) {
                
                // Add the presentation role to each of the navigation bar tabs
                liElem.setAttribute("role", "presentation");
                
                // Add the tab index observer to each of the navigation bar tab links to remove the tabindex attribute whenever it is set or altered
                let aElements = liElem.getElementsByTagName("a");
                if (aElements.length > 0) {
                  tabIndexObserver.observe(aElements[0], {
                    attributes: true,
                    attributeFilter: ["tabindex"]
                  });
                  
                  /*
                    // Set the role of each tab element link to tab
                    // Set aria-selected to true
                    // Set the aria-control attribute equal to the id of the main navbar containing element
                    // Set aria-expanded to true
                    // Set the aria label of the link to explain you can click on the current tab 
                  */
                  for (let aElem of aElements) {
                    aElem.setAttribute("role", "tab");
                    aElem.setAttribute("aria-selected", "true");
                    aElem.setAttribute("aria-controls", "tab-" + navTabID + "-1");
                    aElem.setAttribute("aria-expanded", "true");
                    aElem.setAttribute("aria-label", "Press enter to select the " + aElem.textContent +" tab and show this tab's content below");
                  }
                }
              }
            }
            
            // Set the aria label of each of the accordion dropdown panels in the endpoint modal popup to explain you can click on them to open and close
            let bsCollapses = newNode.getElementsByClassName("panel-group");
            for (let bsCollapse of bsCollapses) {
              let tabInfos = bsCollapse.getElementsByClassName("panel-info");
              for (let tabInfo of tabInfos) {
                tabInfo.setAttribute('aria-label', 'You are currently on a collapsed panel. To open and view the additional information inside, press enter. To close once open, press enter again.');
              }
            }
          }
  
          // Remove the tabindex attribute from the list of fields on the Fields Page
          if (newNode.className === "field-list") {
            let fieldsListTextSection = document.getElementById("fields_page-capstat_fields_text");
            let fieldList = fieldsListTextSection.getElementsByClassName("field-list")[0];
            let ulFieldList = fieldList.getElementsByTagName("ul")[0];
            ulFieldList.removeAttribute("tabindex");
          }
  
          // Remove the tabindex attribute from the list of extensions on the Fields Page
          if (newNode.className === "extension-list") {
            let fieldsListTextSection = document.getElementById("fields_page-capstat_extension_text");
            let fieldList = fieldsListTextSection.getElementsByClassName("extension-list")[0];
            let ulFieldList = fieldList.getElementsByTagName("ul")[0];
            ulFieldList.removeAttribute("tabindex");
          }   
        }
  
        /*
          // Set the aria labels for all the select HTML elements with class shiny-bound-input in the endpoint modal popup 
          // If the added node contains the class 'container-fluid', that means the added node is the endpoint modal popup
          // This sets the aria label of all of the dropdowns in the endpoint modal popup to explain you can use the arrow keys to navigate the filter dropdown
        */
        if (mutation.addedNodes[0].classList && mutation.addedNodes[0].classList.contains("container-fluid")) {
          let containerNode = mutation.addedNodes[0];
          let selectDropdowns = containerNode.querySelectorAll("select.shiny-bound-input");
          for (let selectDropdown of selectDropdowns) {
            selectDropdown.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.');
          }
        } 
      }
  
      // Sets the aria label of the validation table to explain how to use and navigate the validation table to filter the validation failure table
      if (mutation.target.id === "validations_page-validation_details_table") {
        let reactTable = mutation.target.getElementsByClassName("ReactTable");
        let rtTable = reactTable[0].getElementsByClassName("rt-table");
        let rtTHead = rtTable[0].getElementsByClassName("rt-thead");
        let rtTr = rtTHead[0].getElementsByClassName("rt-tr");
        
        let rtTh = rtTr[0].getElementsByClassName("rt-align-left -cursor-pointer rt-th");
        let rtSortHeader = rtTh[0].getElementsByClassName("rt-sort-header");
        let rtThContent = rtSortHeader[0].getElementsByClassName("rt-th-content");
  
        rtThContent[0].setAttribute("aria-label", "You are currently on a table whose entries serve as a filter for the validation failure table. To enter the table, press the tab key. Then, use the up and down arrow keys to move through the filter options. The filter option you are currently focused on will be automatically selected to filter the validation failure table. To exit the filter table, press the tab key again");
      }
  
      // Sets the aria label of the FHIR operations input filter box on the Resource tab to explain how to use and navigate the filter box
      if (mutation.target.classList && mutation.target.classList.contains("selectize-dropdown-content")) {
        let optionElems = mutation.target.getElementsByClassName("option");
        optionElems[0].setAttribute("aria-label", "You are currently in a FHIR operations input filter box. Type to search for an operation, or use the arrow keys to navigate through the operations and select using enter. Remove selected operations by pressing the backspace key. Exit the filter input box by pressing the tab key");
      }
  
      // Checks if the mutated node was the page title to catch when the Lantern tab changes
      if (mutation.target.id === "page_title") {
        
        let pageTitleText = mutation.target.textContent;
        /*
          // Checks if the current page is the Organization, Resource, or Profile page
          // Alter navigation tabs on these pages to have the same structure, classes, and attributes as the navigation bars on the rest of the Lantern pages
        */
        if (pageTitleText === "Organizations Page" || pageTitleText === "Resource Page" || pageTitleText === "Profile Page") {
          
          let navBarTabs = document.getElementsByClassName("nav nav-tabs");
  
          
          let tabbable = document.getElementsByClassName("tabbable");
          let tabContents = tabbable[0].getElementsByClassName("tab-content");
          
  
          for (let navTab of navBarTabs) {
            
            // Add the tablist role to the main navbar containing element
            navTab.setAttribute("role", "tablist");
            
            let navTabID = navTab.getAttribute("data-tabsetid");
            
            // Add the shiny-tab-input and shiny-bound-input classes to the main navbar containing element
            navTab.classList.add("shiny-tab-input", "shiny-bound-input");
  
            let liElements = navTab.getElementsByTagName("li");
            for (let liElem of liElements) {
              
              // Add the presentation role to each of the navigation bar tabs
              liElem.setAttribute("role", "presentation");
              
              // Add the tab index observer to each of the navigation bar tab links to remove the tabindex attribute whenever it is set or altered
              let aElements = liElem.getElementsByTagName("a");                 
              if (aElements.length > 0) {
                tabIndexObserver.observe(aElements[0], {
                  attributes: true,
                  attributeFilter: ["tabindex"]
                });
                  
                /*
                  // Set the role of each tab element link to tab
                  // Set aria-selected to true
                  // Set the aria-control attribute equal to the id of the main navbar containing element
                  // Set aria-expanded to true
                  // Set the aria label of the link to explain you can click on the current tab 
                */
                for (let aElem of aElements) {
                  aElem.setAttribute("role", "tab");
                  aElem.setAttribute("aria-selected", "true");
                  aElem.setAttribute("aria-controls", "tab-" + navTabID + "-1");
                  aElem.setAttribute("aria-expanded", "true");
                  aElem.setAttribute("aria-label", "Press enter to select the " + aElem.textContent +" tab and show this tab's content below");
                }
              }
            }
          }
  
          /*
            // Set the tabindex observer on each of the tab pages within the navigation tabs on the Organization, Resource, or Profile page
            // Which will remove the tab index attribute whenever it is set or altered 
          */
          for (let tabContent of tabContents) {
            let tabPanes = tabContent.getElementsByClassName("tab-pane");
            for (let tabPane of tabPanes) {
              tabIndexObserver.observe(tabPane, {
                  attributes: true,
                  attributeFilter: ["tabindex"]
                });
            }
          }            
        }
  
        if (pageTitleText === "List of Endpoints") {
          let documentation_tab_link = document.getElementById("documentation_page_link");
          documentation_tab_link.addEventListener('click', function (e) {
            let documentation_tab = document.querySelector("a[href = '#shiny-tab-documentation_tab']"); 
            documentation_tab.click();
          });
  
          documentation_tab_link.addEventListener('keyup', function (e) {
            if (event.keyCode === 13) {
              let documentation_tab = document.querySelector("a[href = '#shiny-tab-documentation_tab']"); 
              documentation_tab.click();
            }
          });
        }
        
        /*
          // Set the aria labels for all the select HTML elements with class shiny-bound-input on the Lantern website
          // This sets the aria label of all of the dropdowns on Lantern to explain you can use the arrow keys to navigate the filter dropdown
        */
        let selectInputButtons = document.querySelectorAll("select.shiny-bound-input");
        for (let selectInput of selectInputButtons) {
          selectInput.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.');
        }
        
      }
      
      // Add an aria-label attribute to the FHIR Version dropdown that explains how to navigate and use the filter
      if (mutation.target.id === "show_filters") {
        let dropDownButtons = document.getElementsByClassName("dropdown-toggle");
        for (let dropDownButton of dropDownButtons) {
          dropDownButton.setAttribute('aria-label', 'Dropdown filter menu button. Press the down arrow key to open the filter menu, use the tab or arrow keys to navigate through options, press enter to select a filter option, and use the escape key to close the filter menu.');
        }  
      }
    }
  });
  
  // Set the newNodesObserver on the Lantern document so that it watches for any added elements or any changes to the elements within the overall document
  newNodesObserver.observe(document.body, {
      childList: true, 
      subtree: true, 
      attributes: false, 
      characterData: false
  });