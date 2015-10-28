json.array! @checks do |check|
  json.partial!('/api/private/user_checks/check', check: check)
end
