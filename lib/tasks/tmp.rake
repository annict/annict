# frozen_string_literal: true

namespace :tmp do
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
  end

  task create_records: :environment do
    [EpisodeRecord, WorkRecord].each do |m|
      m.where(record: nil).find_each do |r|
        ActiveRecord::Base.transaction do
          puts "#{r.class.name}: #{r.id}"
          record = Record.create!(user: r.user)
          r.update_column(:record_id, record.id)
        end
      end
    end
  end
end
