# frozen_string_literal: true

namespace :tmp do
  task add_work_id_to_db_activities: :environment do
    DbActivity.find_each do |a|
      case a.trackable_type
      when "Work", "Cast", "Person", "Staff", "Episode", "Organization"
    end
  end
end
