# frozen_string_literal: true

class UpdateForManualPrograms < ActiveRecord::Migration[5.2]
  def change
    change_column :episodes, :raw_number, "float USING raw_number::double precision"

    add_column :channel_groups, :unpublished_at, :datetime
    add_column :program_details, :rebroadcast, :boolean, default: false, null: false
    add_column :programs, :program_detail_id, :integer
    add_column :programs, :number, :integer

    add_index :programs, :program_detail_id
    add_index :programs, %i[program_detail_id episode_id], unique: true
    add_index :programs, %i[program_detail_id number], unique: true
    add_index :programs, %i[program_detail_id started_at], unique: true

    add_foreign_key :programs, :program_details

    change_column_default :channel_groups, :sort_number, 0

    change_column_null :channel_groups, :sc_chgid, true
    change_column_null :channel_groups, :sort_number, false
    change_column_null :programs, :episode_id, true
  end
end
