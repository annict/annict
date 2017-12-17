# frozen_string_literal: true

namespace :tmp do
  task convert_rating_to_rating_state: :environment do
    ActiveRecord::Base.transaction do
      Checkin.where.not(rating: nil).find_each do |r|
        puts "Record #{r.id}: #{r.rating} => #{r.rating_to_rating_state}"
        r.update_column(:rating_state, r.rating_to_rating_state)
      end
    end
  end

  task reset_rating_avg_and_score: :environment do
    Episode.update_all(rating_avg: nil, score: nil)
  end
end
