# frozen_string_literal: true

namespace :episode do
  task update_score: :environment do
    results = {}

    Episode.published.recorded.find_each do |episode|
      rating_states = episode.records.where.not(rating_state: nil).pluck(:rating_state)

      if rating_states.length < 3
        episode.update_columns(rating_avg: nil, satisfaction_score: nil)
        next
      end

      ratings = rating_states.map do |state|
        case state
        when "bad" then -1
        when "average" then 0
        when "good" then 1
        when "great" then 2
        end
      end

      rating_sum = ratings.reduce(:+)
      rating_avg = (rating_sum.to_f / ratings.length).round(1)

      puts "Episode: #{episode.id} => rating_avg: #{rating_avg}"

      results[episode.id] = rating_avg
    end

    rating_max = results.values.max

    results.each do |episode_id, rating_avg|
      satisfaction_score = (rating_avg / rating_max * 100).round(1)
      puts "Episode: #{episode_id} => satisfaction_score: #{satisfaction_score}"
      Episode.update(episode_id, rating_avg: rating_avg, satisfaction_score: satisfaction_score)
    end
  end
end
