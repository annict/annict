# frozen_string_literal: true

namespace :tmp do
  task move_from_collection_to_tag: :environment do
    ActiveRecord::Base.transaction do
      Collection.find_each do |c|
        puts "collection: #{c.id}"
        c.collection_items.each do |ci|
          c.user.add_work_tag!(ci.work, c.title)
        end
      end
    end
  end
end
