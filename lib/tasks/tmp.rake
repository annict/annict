# frozen_string_literal: true

namespace :tmp do
  task update_display_option_user_work_list: :environment do
    ActiveRecord::Base.transaction do
      Setting.
        where(display_option_user_work_list: :list).
        update_all(display_option_user_work_list: :grid_detailed)
      Setting.
        where(display_option_work_list: :list).
        update_all(display_option_work_list: :list_detailed)
    end
  end

  task set_color_rgb_to_work_image: :environment do
    ActiveRecord::Base.transaction do
      WorkImage.find_each do |wi|
        puts "work_image: #{wi.id}"
        wi.update_column(:color_rgb, wi.colors(wi.attachment_url).first) if wi.colors.first.present?
      end
    end
  end
end
