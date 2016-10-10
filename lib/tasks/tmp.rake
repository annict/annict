# frozen_string_literal: true

namespace :tmp do
  task add_work_id_to_db_activities: :environment do
    DbActivity.find_each do |a|
      case a.trackable_type
      when "Work", "Cast", "Person", "Staff", "Episode", "Organization"
      end
    end
  end

  task add_time_zone_to_users: :environment do
    User.find_each do |u|
      puts "Updating user: #{u.id}"
      u.update_column(:time_zone, "Tokyo")
    end
  end

  task create_characters: :environment do
    Cast.find_each do |c|
      puts "Creating character: #{c.part}"
      Character.where(name: c.part).first_or_create!
    end
  end
end
