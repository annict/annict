# frozen_string_literal: true

partial_path = "/api/internal/latest_statuses/latest_status"
json.partial!(partial_path, latest_status: @latest_status)
