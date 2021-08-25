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

target_records = Record.where(migrated_at: nil)
target_records = target_records.where(user_id: user_id) if user_id
target_records = target_records.after(from, field: :updated_at) if from

target_records.preload(:episode_record, :work_record).find_in_batches(batch_size: 2_000) do |records|
  migrated_at = Time.zone.now
  attributes = records.map do |record|
    if record.episode_record?
      episode_record = record.episode_record
      record.attributes.merge(
        "episode_id" => episode_record.episode_id,
        "oauth_application_id" => episode_record.oauth_application_id,
        "body" => episode_record.body.presence || "",
        "comments_count" => episode_record.comments_count,
        "likes_count" => episode_record.likes_count,
        "locale" => new_locale(episode_record.locale),
        "rating" => new_rating(episode_record.rating_state),
        "advanced_rating" => episode_record.rating,
        "watched_at" => episode_record.record.created_at,
        "modified_at" => episode_record.modify_body? ? episode_record.record.updated_at : nil,
        "migrated_at" => migrated_at
      )
    else
      work_record = record.work_record

      unless work_record
        record.destroy!
        next
      end

      record.attributes.merge(
        "oauth_application_id" => work_record.oauth_application_id,
        "body" => work_record.body,
        "likes_count" => work_record.likes_count,
        "locale" => new_locale(work_record.locale),
        "rating" => new_rating(work_record.rating_overall_state),
        "watched_at" => work_record.record.created_at,
        "modified_at" => work_record.modified_at,
        "migrated_at" => migrated_at
      )
    end
  end

  result = Record.upsert_all(attributes.compact)
  puts "upserted: #{result.rows.first(3)}..."
end
