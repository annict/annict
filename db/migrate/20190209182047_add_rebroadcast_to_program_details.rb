# frozen_string_literal: true

class AddRebroadcastToProgramDetails < ActiveRecord::Migration[5.2]
  def change
    add_column :program_details, :rebroadcast, :boolean, default: false, null: false
  end
end
