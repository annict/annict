# frozen_string_literal: true

class FinishUserTipsJob < ApplicationJob
  queue_as :default

  def perform(user, slug)
    UserTipsService.new(user).finish!(slug)
  end
end
