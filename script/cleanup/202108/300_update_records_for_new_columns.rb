# frozen_string_literal: true

opt = OptionParser.new
params = {}
opt.on("-u VAL") { |v| params[:user_id] = v }
opt.on("-f VAL") { |v| params[:from] = v }
opt.parse!(ARGV)

user_id = params[:user_id]
from = params[:from]

def new_locale(old_locale)
  return nil if old_locale.nil?

  case old_locale.to_s
  when "en" then 2
  when "ja" then 1
  when "other" then 0
  else
    raise "Unexpected locale: #{old_locale.inspect}"
  end
end

def new_rating(old_rating)
  return nil if old_rating.nil?

  case old_rating.to_s
  when "bad" then 10
  when "average" then 20
  when "good" then 30
  when "great" then 40
  else
    raise "Unexpected rating: #{old_rating.inspect}"
  end
end

target_work_records = WorkRecord.all
target_work_records = target_work_records.where(user_id: user_id) if user_id
target_work_records = target_work_records.after(from, field: :updated_at) if from

target_work_records.preload(:record).find_each(order: :desc) do |wr|
  p "work_records.id: #{wr.id}"

  wr.record.update_columns(
    oauth_application_id: wr.oauth_application_id,
    body: wr.body,
    likes_count: wr.likes_count,
    locale: new_locale(wr.locale),
    rating: new_rating(wr.rating_overall_state),
    animation_rating: new_rating(wr.rating_animation_state),
    character_rating: new_rating(wr.rating_character_state),
    music_rating: new_rating(wr.rating_music_state),
    story_rating: new_rating(wr.rating_story_state),
    watched_at: wr.record.created_at,
    modified_at: wr.modified_at
  )
end

target_episode_records = EpisodeRecord.all
target_episode_records = target_episode_records.where(user_id: user_id) if user_id
target_episode_records = target_episode_records.after(from, field: :updated_at) if from

target_episode_records.preload(:record).find_each(order: :desc) do |er|
  p "episode_records.id: #{er.id}"

  er.record.update_columns(
    episode_id: er.episode_id,
    oauth_application_id: er.oauth_application_id,
    body: er.body.presence || "",
    comments_count: er.comments_count,
    likes_count: er.likes_count,
    locale: new_locale(er.locale),
    rating: new_rating(er.rating_state),
    advanced_rating: er.rating,
    twitter_url_hash: er.twitter_url_hash,
    facebook_url_hash: er.facebook_url_hash,
    watched_at: er.record.created_at,
    modified_at: er.modify_body? ? er.record.updated_at : nil
  )
end
