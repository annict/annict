# frozen_string_literal: true

namespace :tmp do
  task reset_score: :environment do
    Episode.update_all(score: nil)
  end

  task convert_rating_to_rating_state: :environment do
    ActiveRecord::Base.transaction do
      Checkin.where.not(rating: nil).where(rating_state: nil).find_each do |r|
        puts "Record #{r.id}: #{r.rating} => #{r.rating_to_rating_state}"
        r.update_column(:rating_state, r.rating_to_rating_state)
      end
    end
  end
end
