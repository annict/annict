# frozen_string_literal: true

namespace :episode do
  task update_avg_rating: :environment do
    Episode.published.recorded.find_each do |episode|
      avg_rating = episode.records.avg_rating
      next if avg_rating.blank?
      episode.update_column(:avg_rating, avg_rating)
      puts "episode #{episode.id}"
    end
  end
end
