json.episodes episodes do |e|
  json.id e.id
  json.number e.number
  json.title e.title
  json.workId e.work_id
  json.workTitle e.work.title

  if user.present?
    user_checkins = user.checkins.where(episode_id: e.id)
    json.userCheckinsCount user_checkins.count
  end
end

json.workId work.id
json.isSignedIn user.present?
