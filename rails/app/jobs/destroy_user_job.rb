# typed: false
# frozen_string_literal: true

class DestroyUserJob < ApplicationJob
  queue_as :default

  def perform(user_id)
    User.find(user_id).destroy_in_batches
  end
end
