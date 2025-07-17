# frozen_string_literal: true

Work.where(no_episodes: true).find_each do |work|
  puts "work: #{work.id}"

  if work.episodes.any?
    puts "work: #{work.id} - skipped"
    next
  end

  puts "work: #{work.id} - creating"

  work.episodes.create!(
    raw_number: 1,
    title: "本編"
  )

  puts "work: #{work.id} - created"
end
