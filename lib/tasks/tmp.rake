# frozen_string_literal: true

namespace :tmp do
  task set_counter: :environment do
    Person.find_each do |p|
      puts "person: #{p.id}"
      Person.reset_counters(p.id, :casts_count)
      Person.reset_counters(p.id, :staffs_count)
    end

    Organization.find_each do |o|
      puts "organization: #{o.id}"
      Organization.reset_counters(o.id, :staffs_count)
    end
  end
end
