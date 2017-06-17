# frozen_string_literal: true

namespace :episode do
  task update_score: :environment do
    hash = {}

    Episode.published.recorded.find_each do |episode|
      puts "episode: #{episode.id}"
      rating_states = episode.records.where.not(rating_state: nil).pluck(:rating_state)
      ratings = rating_states.map do |state|
        case state
        when "bad" then -1
        when "average" then 0
        when "good" then 1
        when "great" then 2
        end
      end

      rating_sum = ratings.reduce(:+)

      if rating_sum.present?
        hash[episode.id] = {
          sum: rating_sum
        }
      end
    end

    average = hash.map { |_, v| v[:sum] }.reduce(:+) / hash.length

    hash.each do |key, val|
      # 平均との差
      hash[key][:abs] = (val[:sum] - average).abs
      # 平方数
      hash[key][:square] = hash[key][:abs] * hash[key][:abs]
    end

    # 分散
    variance = hash.map { |_, v| v[:square] }.reduce(:+) / hash.length
    # 標準偏差
    sqrt = Math.sqrt(variance)

    hash.each do |key, val|
      # 偏差値
      score = ((hash[key][:sum] - average) / sqrt) * 10 + 50
      Episode.update(key, score: score)
    end
  end
end
