# frozen_string_literal: true

namespace :tmp do
  task update_root_resource_on_db_activities: :environment do
    DbActivity.where(trackable_type: %w(Cast Episode Staff Program)).find_each do |a|
      next if a.trackable.blank?
      puts "Activity: #{a.id}"
      a.root_resource = a.trackable.work
      a.save!
    end
  end

  task add_time_zone_to_users: :environment do
    User.find_each do |u|
      puts "Updating user: #{u.id}"
      u.update_column(:time_zone, "Asia/Tokyo")
    end
  end

  task create_characters: :environment do
    Cast.find_each do |c|
      puts "Creating character: #{c.part}"
      ActiveRecord::Base.transaction do
        character = Character.where(name: c.part).first_or_create!
        if c.character_id.blank? && character.name != "-"
          c.update_column(:character_id, character.id)
        end
      end
    end
  end

  task delete_edit_request_records_from_db_activities: :environment do
    DbActivity.where(trackable_type: %w(EditRequest EditRequestComment)).delete_all
  end

  task convert_item_image_to_work_image: :environment do
    Item.find_each do |i|
      next if i.work.work_image.present?
      image_path = i.tombo_image.path(:original)
      image_url = if Rails.env.production?
        "https://d3a8d1smk6xli.cloudfront.net/#{image_path}"
      else
        "http://annict.dev:3000#{image_path}"
      end
      work_image = i.work.create_work_image!(
        user: User.find(2),
        attachment: URI.parse(image_url),
        asin: i.url
      )
      i.work.update_column(:work_image_id, work_image.id)
      puts "Item #{i.id} converted to Work Image #{work_image.id}"
    end
  end

  task set_forum_categories: :environment do
    data = [
      {
        slug: "site_news",
        name: "お知らせ",
        name_en: "Site News",
        description: "サイトに関するお知らせ。",
        description_en: "Announcements about Annict.",
        postable_role: "admin",
        sort_number: 0
      },
      {
        slug: "general",
        name: "一般",
        name_en: "General",
        description: "カテゴリ分けできない一般的な話題。",
        description_en: "General discussions which can't be categorize.",
        postable_role: "user",
        sort_number: 10
      },
      {
        slug: "feedback",
        name: "フィードバック",
        name_en: "Feedback",
        description: "Annictに改善要望等のフィードバックがありましたらこちらへどうぞ。",
        description_en: "Have an idea or suggestion for the site? Share it here.",
        postable_role: "user",
        sort_number: 20
      },
      {
        slug: "db_request",
        name: "データ編集リクエスト",
        name_en: "DB Modification Requests",
        description: "Annictに登録されているデータに不備等がありましたら、こちらにお願いします。",
        description_en: "Have a request for anime data? Share it here.",
        postable_role: "user",
        sort_number: 30
      }
    ]

    data.each { |d| ForumCategory.create!(d) }
  end
end
