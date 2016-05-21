# frozen_string_literal: true

json.array! @latest_statuses do |latest_status|
  partial_path = "/api/internal/latest_statuses/latest_status"
  json.partial!(partial_path, latest_status: latest_status)
end
