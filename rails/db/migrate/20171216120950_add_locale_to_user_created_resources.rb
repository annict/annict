# frozen_string_literal: true

class AddLocaleToUserCreatedResources < ActiveRecord::Migration[5.1]
  def change
    table_names = %i[
      checkins
      comments
      db_comments
      forum_comments
      forum_posts
      items
      reviews
      tips
      userland_projects
      work_comments
      work_taggables
      work_tags
    ]

    table_names.each do |tn|
      add_column tn, :locale, :string, null: false, default: "other"
      add_index tn, :locale
    end
  end
end
