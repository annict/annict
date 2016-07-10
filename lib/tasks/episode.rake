namespace :episode do
  task update_avg_rating: :environment do
    Episode.published.recorded.find_each do |episode|
      ratings = episode.checkins.pluck(:rating).select(&:present?)
      next if ratings.blank?

      avg_rating = (ratings.inject { |sum, rating| sum + rating } / ratings.count).round(1)
      episode.update_column(:avg_rating, avg_rating)
      puts "episode #{episode.id}"
    end
  end
end
