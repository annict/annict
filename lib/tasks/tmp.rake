# frozen_string_literal: true

namespace :tmp do
  task convert_activity_actions: :environment do
    def update_activity(a)
      puts "activity id: #{a.id}"

      case a.action
      when "statuses.create"
        a.update_column :action, "create_status"
      when "checkins.create"
        a.update_column :action, "create_record"
      end
    end

    Activity.order(id: :desc).limit(3000).each do |a|
      update_activity a
    end

    Activity.find_each do |a|
      update_activity a
    end
  end
end
