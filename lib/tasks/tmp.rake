# frozen_string_literal: true

namespace :tmp do
  task update_status_id_on_library_entries: :environment do
    LibraryEntry.by_month(field: :updated_at).find_each do |ls|
      puts "id: #{ls.id}"
      status = Status.where(user_id: ls.user_id, work_id: ls.work_id, kind: ls.kind.value).order(id: :desc).first
      ls.update_column(:status_id, status.id)
    end
  end
end
