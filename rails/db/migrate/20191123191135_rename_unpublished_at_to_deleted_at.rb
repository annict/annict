# frozen_string_literal: true

class RenameUnpublishedAtToDeletedAt < ActiveRecord::Migration[6.0]
  def change
    rename_column :channel_groups, :unpublished_at, :deleted_at
  end
end
