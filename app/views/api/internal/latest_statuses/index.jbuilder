# frozen_string_literal: true

json.latest_statuses @latest_statuses do |latest_status|
  json.partial!("/api/internal/latest_statuses/latest_status", latest_status: latest_status)
end

json.user do
  json.authorized_to_twitter current_user.authorized_to?(:twitter)
  json.authorized_to_facebook current_user.authorized_to?(:facebook)
  json.share_record_to_twitter current_user.setting&.share_record_to_twitter?
  json.share_record_to_facebook current_user.setting&.share_record_to_facebook?
end
