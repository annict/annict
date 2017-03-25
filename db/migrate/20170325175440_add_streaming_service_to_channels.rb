# frozen_string_literal: true

class AddStreamingServiceToChannels < ActiveRecord::Migration[5.0]
  def change
    add_column :channels, :streaming_service, :boolean, default: false
    add_column :channels, :aasm_state, :string, null: false, default: "published"

    add_index :channels, :streaming_service

    change_column_null :channels, :channel_group_id, true
    change_column_null :channels, :sc_chid, true

    add_presence_constraint :channels, :name

    add_inclusion_constraint :channels, :aasm_state, in: %w(published hidden)

    remove_column :channels, :published
  end
end
