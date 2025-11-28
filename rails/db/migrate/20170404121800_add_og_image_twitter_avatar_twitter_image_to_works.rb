# frozen_string_literal: true

class AddOgImageTwitterAvatarTwitterImageToWorks < ActiveRecord::Migration[5.0]
  def change
    add_column :works, :facebook_og_image_url, :string, null: false, default: ""
    add_column :works, :twitter_image_url, :string, null: false, default: ""
    add_column :works, :recommended_image_url, :string, null: false, default: ""
  end
end
