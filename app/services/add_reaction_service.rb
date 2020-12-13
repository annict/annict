# frozen_string_literal: true

class AddReactionService < ApplicationService
  class AddReactionServiceResult < Result::Service
    attr_accessor :reaction
  end

  def initialize(user:, resource:, content: :heart)
    super()
    @user = user
    @resource = resource
    @content = content
  end

  def call
    like = @user.likes.find_by_resource(@resource)

    if like
      @result.reaction = like
      return @result
    end

    ActiveRecord::Base.transaction do
      @result.reaction = @user.add_reaction(@resource, content: @content)
      send_notification
    end

    @result
  end

  private

  def send_notification
    @result.reaction.send_notification_to(@user)
  end

  def result_class
    AddReactionServiceResult
  end
end
