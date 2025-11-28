# frozen_string_literal: true

class AddEventFavoriteWorksAddedToEmailNotifications < ActiveRecord::Migration[5.0]
  def change
    add_column :email_notifications, :event_favorite_works_added, :boolean,
      null: false,
      default: true
  end
end
