# frozen_string_literal: true

namespace :tmp do
  task convert_activity_actions: :environment do
    Activity.find_each do |a|
      puts "activity id: #{a.id}"

      case a.action
      when "statuses.create"
        a.update_column :action, "create_status"
      when "checkins.create"
        a.update_column :action, "create_record"
      end
    end
  end
end
