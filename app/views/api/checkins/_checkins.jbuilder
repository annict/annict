json.array! @checkins do |checkin|
  json.partial!('/api/checkins/checkin', checkin: checkin)
end
