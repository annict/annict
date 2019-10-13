# frozen_string_literal: true

class ChangeSchema20191014 < ActiveRecord::Migration[6.0]
  def change
    enable_extension "citext"

    rename_table :pvs, :trailers

    add_column :episodes, :number_en, :string, null: false, default: ""

    change_column :users, :username, :citext
    change_column :users, :email, :citext

    rename_column :episodes, :episode_records_with_body_count, :episode_record_bodies_count

    rename_column :episode_records, :comment, :body

    rename_column :characters, :favorite_characters_count, :favorite_users_count

    rename_column :organizations, :favorite_organizations_count, :favorite_users_count

    rename_column :people, :favorite_people_count, :favorite_users_count

    remove_index :programs, name: :index_programs_on_program_detail_id_and_started_at
  end
end
