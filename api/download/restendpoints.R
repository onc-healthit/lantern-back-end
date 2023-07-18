library(plumber)


#* @apiTitle Simple API

#* Echo provided text
#* @get /api/download
function(res) {
    res$setHeader("Content-Type", "text/csv")
    res$setHeader("Content-Disposition", "attachment; filename=example.csv")

    # Create a sample data frame
    data <- data.frame(
      Name = c("John", "Jane", "Alice"),
      Age = c(25, 30, 27)
    )

    # Convert the data frame to CSV format
    #csv_data <- utils::write.table(data, sep = ",", quote = FALSE, row.names = FALSE)
    #res$body <- csv_data

    #res.write(csv_data)
    # Return the response
    #return(res)
    write.csv(data,file='example.csv', row.names=FALSE)
    include_file('example.csv', res, content_type = "text/csv")
}

