# frozen_string_literal: true

namespace :tmp do
  task update_records: :environment do
    Checkin.where(work_id: nil).each do |c|
      c.update_column(:work_id, c.episode.work_id)
    end
  end
end
