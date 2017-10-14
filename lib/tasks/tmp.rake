# frozen_string_literal: true

namespace :tmp do
  task set_started_on_and_ended_on_to_works: :environment do
    ActiveRecord::Base.transaction do
      url = "https://github.com/anilogia/animedb/raw/master/animedb.yml"
      content = open(url) { |f| f.read }
      yaml_data = YAML::load(content).select { |data| data.key?("annict_id") }
      yaml_data.each do |data|
        work = Work.find_by(id: data["annict_id"])
        next if work.blank?
        puts "Work: #{work.id}"
        started_on = Date.new(data["started_year"], data["started_month"], data["started_day"]) rescue nil
        ended_on = Date.new(data["ended_year"], data["ended_month"], data["ended_day"]) rescue nil
        work.started_on = started_on unless started_on.nil?
        work.ended_on = ended_on unless ended_on.nil?
        work.save!
      end
    end
  end
end
