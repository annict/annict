json.array! @latest_statuses do |latest_status|
  json.partial!("/api/latest_statuses/latest_status", latest_status: latest_status)
end
