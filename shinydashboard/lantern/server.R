# Define server function
function(input, output, session) {

  callModule(
    dashboard,
    "dashboard_page")
        
   callModule(
    availability,
    "availability_page")
        
   callModule(
    performance,
    "performance_page") 
}