# frozen_string_literal: true

namespace :tmp do
  task :update_foreign_keys_on_activities, %i(from) => :environment do |_, args|
    activity_id = args[:from]

    Activity.where("id >= ?", activity_id).find_each do |a|
      puts "activity: #{a.id}"

      begin
        case a.action
        when "create_status"
          a.update_columns(work_id: a.recipient.id, status_id: a.trackable.id)
        when "create_record"
          a.update_columns(work_id: a.recipient.work.id, episode_id: a.recipient.id, record_id: a.trackable.id)
        when "create_review"
          a.update_columns(work_id: a.recipient.id, review_id: a.trackable.id)
        when "create_multiple_records"
          a.update_columns(work_id: a.recipient.id, multiple_record_id: a.trackable.id)
        end
      rescue => ex
        puts ex
        a.destroy
      end
    end
  end
end
