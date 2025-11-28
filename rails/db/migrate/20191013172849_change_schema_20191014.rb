# frozen_string_literal: true

class ChangeSchema20191014 < ActiveRecord::Migration[6.0]
  def change
    enable_extension "citext"

    rename_table :pvs, :trailers
    rename_table :programs, :slots
    rename_table :program_details, :programs

    remove_column :casts, :part
    remove_column :characters, :kind

    rename_column :characters, :favorite_characters_count, :favorite_users_count
    rename_column :episodes, :episode_records_with_body_count, :episode_record_bodies_count
    rename_column :episode_records, :comment, :body
    rename_column :episode_records, :modify_comment, :modify_body
    rename_column :organizations, :favorite_organizations_count, :favorite_users_count
    rename_column :people, :favorite_people_count, :favorite_users_count
    rename_column :settings, :hide_record_comment, :hide_record_body
    rename_column :settings, :programs_sort_type, :slots_sort_type
    rename_column :slots, :program_detail_id, :program_id

    add_column :channels, :sort_number, :integer, null: false, default: 0
    add_column :episodes, :number_en, :string, null: false, default: ""

    add_index :channels, :sort_number

    change_column :users, :username, :citext
    change_column :users, :email, :citext
  end
end
