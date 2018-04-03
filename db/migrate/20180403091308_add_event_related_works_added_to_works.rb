# frozen_string_literal: true

class AddEventRelatedWorksAddedToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :email_notifications, :event_related_works_added, :boolean,
      null: false,
      default: true
  end
end
