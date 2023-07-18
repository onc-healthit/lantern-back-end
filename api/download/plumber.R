  library(plumber)

  # Create and run the plumber API
  api <- plumber::plumb("restendpoints.R")
  api$run(port = 8989, host = "0.0.0.0")