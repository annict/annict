# frozen_string_literal: true

namespace :tmp do
  task update_status_id_on_latest_statuses: :environment do
    LatestStatus.where(status_id: nil).find_each do |ls|
      puts "id: #{ls.id}"
      status = Status.where(user_id: ls.user_id, work_id: ls.work_id, kind: ls.kind.value).order(id: :desc).first
      ls.update_column(:status_id, status.id)
    end
  end
end
