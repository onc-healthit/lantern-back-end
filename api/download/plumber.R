  library(plumber)

  # Create and run the plumber API
  api <- plumber::plumb("download/restendpoints.R")
  api$run(port = 8989, host = "0.0.0.0")