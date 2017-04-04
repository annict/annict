# frozen_string_literal: true

namespace :tmp do
  task set_year_and_season: :environment do
    Work.find_each do |work|
      if work.season_model.present?
        puts "work: #{work.id}"
        attrs = {
          season_year: work.season_model.year,
          season_name: work.season_model.name
        }
        work.update_attributes(attrs)
      end
    end
  end
end
