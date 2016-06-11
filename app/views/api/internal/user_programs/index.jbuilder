# frozen_string_literal: true

json.programs @programs do |program|
  json.call(program, :rebroadcast)
  json.started_at program.started_at.in_time_zone("Asia/Tokyo").strftime("%m/%d %H:%M")
  json.broadcasted program.broadcasted?

  json.channel do
    json.name program.channel.name
  end

  json.episode do
    json.id program.episode.id
    json.number program.episode.number
    json.title program.episode.title
  end

  json.work do
    json.id program.work.id
    json.title program.episode.work.title
    json.image_url annict_image_url(program.work.item, :tombo_image, size: "48x48")
  end
end

json.user do
  json.authorized_to_twitter current_user.authorized_to?(:twitter)
  json.authorized_to_facebook current_user.authorized_to?(:facebook)
  json.share_record_to_twitter current_user.setting.share_record_to_twitter?
  json.share_record_to_facebook current_user.setting.share_record_to_facebook?
end
