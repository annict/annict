# frozen_string_literal: true

namespace :data_migration do
  task update_users: :environment do
    User.hidden.find_each do |u|
      puts "Update deleted_at on users - user: #{u.id}"
      u.update_column(:deleted_at, u.updated_at)
    end
  end
end
