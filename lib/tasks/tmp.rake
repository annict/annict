# frozen_string_literal: true

namespace :tmp do
  task set_rating_state: :environment do
    Checkin.where.not(rating: nil).find_each do |record|
      record.update_column(:rating_state, record.rating_to_rating_state)
      puts "record: #{record.id} - #{record.rating} -> #{record.rating_state}"
    end
  end
end
