# frozen_string_literal: true

namespace :tmp do
  task move_title_ro_to_title_en: :environment do
    Work.where.not(title_ro: "").find_each do |w|
      puts "Work: #{w.id}"
      w.update_column(:title_en, w.title_ro)
    end

    Episode.where.not(title_ro: "").find_each do |e|
      puts "Episode: #{e.id}"
      e.update_column(:title_en, e.title_ro)
    end
  end
end
