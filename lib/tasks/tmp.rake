# frozen_string_literal: true

namespace :tmp do
  task create_new_records: :environment do
    [Record, Review].each do |m|
      m.where(new_record_id: nil).find_each do |r|
        ActiveRecord::Base.transaction do
          puts "#{r.class.name}: #{r.id}"
          new_record = NewRecord.create!(
            user_id: r.user.id,
            work_id: r.work.id,
            impressions_count: r.impressions_count,
            created_at: r.created_at,
            updated_at: r.updated_at
          )
          r.update_column(:new_record_id, new_record.id)
        end
      end
    end
  end
end
