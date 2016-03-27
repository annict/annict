# frozen_string_literal: true

json.activities @activities do |activity|
  json.partial!("/api/internal/activities/activity", activity: activity)
end
