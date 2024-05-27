# typed: false
# frozen_string_literal: true

namespace :favorite do
  task set_watched_counter: :environment do
    User.find_each do |u|
      puts "user: #{u.id}"
      u.update_watched_works_count
    end
  end
end
