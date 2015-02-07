json.activities @activities do |activity|
  json.partial!('/api/activities/activity', activity: activity)
end
