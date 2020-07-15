library(lubridate)

get_avg_response_time <- function() {
  # get time series of response time metrics for all endpoints
  # groups response time averages by 23 hour intervals and shows data for a range of 30 days
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql("SELECT date.datetime AS time, AVG(fhir_endpoints_info_history.response_time_seconds)
                FROM fhir_endpoints_info_history, (SELECT floor(extract(epoch from fhir_endpoints_info_history.entered_at)/82800)*82800 AS datetime FROM fhir_endpoints_info_history) as date
                GROUP BY time HAVING date.datetime between(date.datetime+0) AND (date.datetime+2592000)
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