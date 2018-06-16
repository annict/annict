# frozen_string_literal: true

namespace :tmp do
  task create_records: :environment do
    [EpisodeRecord, WorkRecord].each do |m|
      m.where(record_id: nil).find_each do |r|
        ActiveRecord::Base.transaction do
          puts "#{r.class.name}: #{r.id}"
          attrs = {
            user_id: r.user.id,
            work_id: r.work.id,
            created_at: r.created_at,
            updated_at: r.updated_at
          }
          attrs[:impressions_count] = r.impressions_count if r.instance_of?(WorkRecord)
          record = Record.create!(attrs)
          r.update_column(:record_id, record.id)
        end
      end
    end
  end

  task rename_record_types: :environment do
    Activity.where(trackable_type: "Checkin").update_all(trackable_type: "EpisodeRecord")
    Activity.where(trackable_type: "Record").update_all(trackable_type: "EpisodeRecord")
    Activity.where(trackable_type: "Review").update_all(trackable_type: "WorkRecord")
    Activity.where(trackable_type: "MultipleRecord").update_all(trackable_type: "MultipleEpisodeRecord")

    Activity.where(action: "create_record").update_all(action: "create_episode_record")
    Activity.where(action: "create_review").update_all(action: "create_work_record")
    Activity.where(action: "create_multiple_records").update_all(action: "create_multiple_episode_records")

    Like.where(recipient_type: "Checkin").update_all(recipient_type: "EpisodeRecord")
    Like.where(recipient_type: "Record").update_all(recipient_type: "EpisodeRecord")
    Like.where(recipient_type: "Review").update_all(recipient_type: "WorkRecord")
    Like.where(recipient_type: "MultipleRecord").update_all(recipient_type: "MultipleEpisodeRecord")

    Setting.where(display_option_record_list: "my_records").update_all(display_option_record_list: "my_episode_records")

    ActiveRecord::Base.transaction do
      Impression.find_each do |i|
        puts i.id
        case i.impressionable_type
        when "Review"
          i.update(
            impressionable_type: "Record",
            impressionable_id: WorkRecord.find(i.impressionable_id).record_id,
            controller_name: "records"
          )
        when "Record"
          next if i.impressionable.episode_record.present? || i.impressionable.work_record.present?
          i.update(impressionable_id: EpisodeRecord.find(i.impressionable_id).record_id)
        end
      end
    end
  end
end
