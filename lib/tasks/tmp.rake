# frozen_string_literal: true

namespace :tmp do
  task create_new_records: :environment do
    [Record, Review].each do |m|
      m.where(new_record_id: nil).find_each do |r|
        ActiveRecord::Base.transaction do
          puts "#{r.class.name}: #{r.id}"
          attrs = {
            user_id: r.user.id,
            work_id: r.work.id,
            created_at: r.created_at,
            updated_at: r.updated_at
          }
          attrs[:impressions_count] = r.impressions_count if r.instance_of?(Review)
          new_record = NewRecord.create!(attrs)
          r.update_column(:new_record_id, new_record.id)
        end
      end
    end
  end
end
