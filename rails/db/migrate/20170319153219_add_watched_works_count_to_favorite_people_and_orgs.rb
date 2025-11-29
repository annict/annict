# frozen_string_literal: true

class AddWatchedWorksCountToFavoritePeopleAndOrgs < ActiveRecord::Migration[5.0]
  def change
    add_column :favorite_people, :watched_works_count, :integer,
      null: false,
      default: 0
    add_column :favorite_organizations, :watched_works_count, :integer,
      null: false,
      default: 0

    add_index :favorite_people, :watched_works_count
    add_index :favorite_organizations, :watched_works_count
  end
end
