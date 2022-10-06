# frozen_string_literal: true

Work.where(no_episodes: true).find_each do |w|
  next if w.episodes.count >= 1

  puts "work: #{w.id}"

  w.episodes.create!(title: w.title)
end
