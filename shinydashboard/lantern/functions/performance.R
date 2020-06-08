library(lubridate)

get_avg_response_time <- function() {
  # get time series of response time metrics for all endpoints
  # will update with dynamic time ranges, group by 4 minute intervals
  all_endpoints_response_time <- as_tibble(tbl(db_connection,sql("SELECT floor(extract(epoch from metrics_values.time)/240)*240 AS time, AVG(metrics_values.value) 
  FROM metrics_labels, metrics_values
  WHERE metrics_labels.metric_name = 'AllEndpoints_http_response_time' 
  AND metrics_labels.id = metrics_values.labels_id
  AND metrics_values.time BETWEEN '2020-01-01T00:00:00Z' AND '2020-07-01T00:00:00Z'
  GROUP BY time
  ORDER BY time ")) ) %>% mutate(date=as_datetime(time)) %>% select(date,avg)

  # convert to xts format for use in dygraph
  xts(x = all_endpoints_response_time$avg, order.by = all_endpoints_response_time$date)
}