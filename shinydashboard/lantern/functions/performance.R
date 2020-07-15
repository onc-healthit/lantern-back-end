library(lubridate)

get_avg_response_time <- function() {
  # get time series of response time metrics for all endpoints
  # groups by 4 minute intervals
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql("SELECT floor(extract(epoch from fhir_endpoints_info_history.entered_at)/240)*240 AS time, AVG(fhir_endpoints_info_history.response_time_seconds)
            FROM fhir_endpoints_info_history
            WHERE fhir_endpoints_info_history.entered_at BETWEEN '2020-01-01T00:00:00Z' AND '2020-08-01T00:00:00Z'
            GROUP BY time
            ORDER BY time")
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, avg)

  # convert to xts format for use in dygraph
  xts(x = all_endpoints_response_time$avg,
      order.by = all_endpoints_response_time$date
  )
}