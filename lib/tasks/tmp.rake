# frozen_string_literal: true

namespace :tmp do
  task create_records: :environment do
    [EpisodeRecord, WorkRecord].each do |m|
      m.where(record: nil).find_each do |r|
        ActiveRecord::Base.transaction do
          puts "#{r.class.name}: #{r.id}"
          record = r.user.records.create!(
            work: r.work,
            impressions_count: r.impressions_count,
            created_at: r.created_at,
            updated_at: r.updated_at
          )
          r.update_column(:record_id, record.id)
        end
      end
    end
  end

  task rename_record_types: :environment do
    Activity.where(trackable_type: "Checkin").update_all(trackable_type: "EpisodeRecord")
    Activity.where(trackable_type: "Review").update_all(trackable_type: "WorkRecord")
    Activity.where(trackable_type: "MultipleRecord").update_all(trackable_type: "MultipleEpisodeRecord")

    Activity.where(action: "create_record").update_all(action: "create_episode_record")
    Activity.where(action: "create_review").update_all(action: "create_work_record")
    Activity.where(action: "create_multiple_records").update_all(action: "create_multiple_episode_records")

    Like.where(recipient_type: "Checkin").update_all(recipient_type: "EpisodeRecord")
    Like.where(recipient_type: "Review").update_all(recipient_type: "WorkRecord")
    Like.where(recipient_type: "MultipleRecord").update_all(recipient_type: "MultipleEpisodeRecord")

    ActiveRecord::Base.transaction do
      Impression.find_each do |i|
        case i.impressionable_type
        when "Review"
          i.update(
            impressionable_type: "Record",
            impressionable_id: WorkRecord.find(i.impressionable_id).record_id,
            controller_name: "records"
          )
        when "Record"
          i.update(impressionable_id: EpisodeRecord.find(i.impressionable_id).record_id)
        end
      end
    end
  end

  task update_records: :environment do
    Record.find_each do |r|
      puts r.id
      resource_record = r.episode_record.presence || r.work_record
      r.update_column(:work_id, resource_record.work_id)
    end
  end
end
