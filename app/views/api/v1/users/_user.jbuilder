# frozen_string_literal: true

json.id user.id if @params.fields_contain?("#{field_prefix}id")
json.username user.username if @params.fields_contain?("#{field_prefix}username")
json.name user.profile.name if @params.fields_contain?("#{field_prefix}name")
json.description user.profile.description if @params.fields_contain?("#{field_prefix}description")
json.url user.profile.url if @params.fields_contain?("#{field_prefix}url")
json.records_count user.checkins_count if @params.fields_contain?("#{field_prefix}records_count")
json.created_at user.created_at if @params.fields_contain?("#{field_prefix}created_at")

if show_all
  json.email user.email if @params.fields_contain?("#{field_prefix}email")
  json.notifications_count user.notifications_count if @params.fields_contain?("#{field_prefix}notifications_count")
end
