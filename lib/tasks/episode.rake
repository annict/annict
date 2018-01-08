# frozen_string_literal: true

namespace :episode do
  task update_score: :environment do
    RATING_MAX = 2

    works = Work.published
    watchers_count_max = works.pluck(:watchers_count).max

    works.find_each do |work|
      watchers_count = work.watchers_count
      watchers_rate = 1 + (watchers_count.to_f / watchers_count_max) * 10

      episodes = work.episodes.published.recorded
      next if episodes.blank?

      episodes.find_each do |episode|
        rating_states = episode.records.where.not(rating_state: nil).pluck(:rating_state)

        if rating_states.length.zero?
          episode.update_columns(score: nil)
          next
        end

        ratings = rating_states.map do |state|
          case state.to_s
          when "bad" then 0
          when "average" then 1
          when "good" then 2
          when "great" then [2, 2]
          end
        end.flatten

        ratings_count = ratings.length
        ratings_sum = ratings.inject(:+)
        ratings_avg = (ratings_sum.to_f / ratings_count).round(1)
        total = ratings_count * watchers_rate
        rating_range = 0..RATING_MAX
        wilson_score = WilsonScore.rating_lower_bound(ratings_avg, total, rating_range)
        score = (wilson_score / RATING_MAX * 10).round(2)

        outputs = [
          "ratings_count: #{ratings_count}",
          "rating_sum: #{ratings_sum}",
          "rating_avg: #{ratings_avg}",
          "total: #{total}",
          "wilson_score: #{wilson_score}",
          "score: #{score}"
        ]
        puts "Work: #{work.id}, Episode: #{episode.id} => #{outputs.join(', ')}"

        episode.update_column(:score, score)
      end
    end
  end
end
