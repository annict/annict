# frozen_string_literal: true

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

WorkRecord.where(migrated_at: nil).find_in_batches(batch_size: 2_000) do |work_records|
  migrated_at = Time.zone.now
  attributes = work_records.map do |work_record|
    work_record.attributes.merge(
      "animation_rating" => new_rating(work_record.rating_animation_state),
      "character_rating" => new_rating(work_record.rating_character_state),
      "music_rating" => new_rating(work_record.rating_music_state),
      "story_rating" => new_rating(work_record.rating_story_state),
      "migrated_at" => migrated_at
    )
  end

  result = WorkRecord.upsert_all(attributes.compact)
  puts "upserted: #{result.rows.first(3)}..."
end
