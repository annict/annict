# frozen_string_literal: true

class ChangeGumroadSubscribers202010 < ActiveRecord::Migration[6.0]
  def change
    change_column_null :gumroad_subscribers, :gumroad_user_id, true
    change_column_null :gumroad_subscribers, :gumroad_user_email, true
  end
end
