json.activities @activities do |activity|
  json.partial!('/api/private/activities/activity', activity: activity)
end
