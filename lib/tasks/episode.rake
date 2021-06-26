# frozen_string_literal: true

namespace :episode do
  task generate: :environment do
    EpisodeGeneratorService.execute!
    SyobocalEpisodeDataFetcherService.execute!
  end

  task update_score: :environment do
    RATING_MAX = 2

    works = Anime.only_kept

    works.find_each do |work|
      episodes = work.episodes.only_kept.recorded
      next if episodes.blank?

      episodes.find_each do |episode|
        rating_states = episode.episode_records.only_kept.where.not(rating_state: nil).pluck(:rating_state)

        if rating_states.length.zero?
          episode.update_columns(satisfaction_rate: nil)
          next
        end

        ratings = rating_states.map { |state|
          case state.to_s
          when "bad" then 0
          when "average" then 1
          when "good", "great" then RATING_MAX
          end
        }

        ratings_count = ratings.length
        ratings_sum = ratings.inject(:+)
        ratings_avg = ratings_sum.to_f / ratings_count
        satisfaction_rate = (ratings_avg / RATING_MAX * 100).round(2)

        outputs = [
          "ratings_count: #{ratings_count}",
          "ratings_sum: #{ratings_sum}",
          "ratings_avg: #{ratings_avg}",
          "satisfaction_rate: #{satisfaction_rate}"
        ]
        puts "Work: #{work.id}, Episode: #{episode.id} => #{outputs.join(", ")}"

        episode.update_columns(satisfaction_rate: satisfaction_rate, ratings_count: ratings_count)
      end
    end
  end
end
