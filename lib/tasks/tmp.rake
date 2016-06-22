# frozen_string_literal: true

namespace :tmp do
  task update_item_color: :environment do
    Item.find_each do |i|
      i.update_column(:image_color_light, i.image_colors[:light])
      i.update_column(:image_color_dark, i.image_colors[:dark])
      puts "updated: #{i.id}"
    end
  end
end
