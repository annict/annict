json.array! @checks do |check|
  json.partial!('/api/user_checks/check', check: check)
end
