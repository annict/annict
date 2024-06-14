# frozen_string_literal: true

opt = OptionParser.new
params = {}
opt.on("-u VAL") { |v| params[:user_id] = v }
opt.on("-f VAL") { |v| params[:from] = v }
opt.parse!(ARGV)

user_id = params[:user_id]
from = params[:from]

target_records = Record.all
target_records = target_records.where(user_id:) if user_id
target_records = target_records.after(from, field: :updated_at) if from

target_records.preload(:episode_record, :work_record).find_in_batches(batch_size: 2_000) do |records|
  attributes = records.map do |record|
    if record.episode_record?
      episode_record = record.episode_record
      record.attributes.merge(
        "trackable_id" => episode_record.episode_id,
        "trackable_type" => "Episode",
        "oauth_application_id" => episode_record.oauth_application_id,
        "body" => episode_record.body.presence || "",
        "comments_count" => episode_record.comments_count,
        "likes_count" => episode_record.likes_count,
        "locale" => episode_record.locale,
        "overall_rating" => episode_record.rating_state,
        "advanced_overall_rating" => episode_record.rating,
        "modified_at" => episode_record.modify_body? ? episode_record.record.updated_at : nil
      )
    else
      work_record = record.work_record

      unless work_record
        record.destroy!
        next
      end

      record.attributes.merge(
        "trackable_id" => work_record.work_id,
        "trackable_type" => "Work",
        "oauth_application_id" => work_record.oauth_application_id,
        "body" => work_record.body,
        "likes_count" => work_record.likes_count,
        "locale" => work_record.locale,
        "overall_rating" => work_record.rating_overall_state,
        "animation_rating" => work_record.rating_animation_state,
        "character_rating" => work_record.rating_character_state,
        "music_rating" => work_record.rating_music_state,
        "story_rating" => work_record.rating_story_state,
        "modified_at" => work_record.modified_at
      )
    end
  end

  result = Record.upsert_all(attributes.compact)
  puts "Upserted: #{result.rows.first(3)}..."
end
