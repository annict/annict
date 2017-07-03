# frozen_string_literal: true

class ByeByeJob < ApplicationJob
  queue_as :default

  def perform(user_id)
    user = User.find(user_id)

    ActiveRecord::Base.transaction do
      user.destroy
      user.oauth_applications.available.find_each do |app|
        app.update(owner: nil)
        app.hide!
      end
    end
  end
end
