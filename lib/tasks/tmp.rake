# frozen_string_literal: true

namespace :tmp do
  task update_root_resource_on_db_activities: :environment do
    DbActivity.where(trackable_type: %w(Cast Episode Staff Program)).find_each do |a|
      next if a.trackable.blank?
      puts "Activity: #{a.id}"
      a.root_resource = a.trackable.work
      a.save!
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
      ActiveRecord::Base.transaction do
        character = Character.where(name: c.part).first_or_create!
        c.update_column(:character_id, character.id) if character.name != "-"
      end
    end
  end

  task delete_edit_request_records_from_db_activities: :environment do
    DbActivity.where(trackable_type: %w(EditRequest EditRequestComment)).delete_all
  end
end
