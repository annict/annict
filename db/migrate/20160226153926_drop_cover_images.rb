# frozen_string_literal: true

class DropCoverImages < ActiveRecord::Migration
  def change
    drop_table :cover_images
  end
end
