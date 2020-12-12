# frozen_string_literal: true

class AddReactionService < ApplicationService
  class AddReactionServiceResult < ServiceResult
    attr_accessor :reaction
  end

  def initialize(user:, resource:)
    super()
    @user = user
    @resource = resource
  end

  def call
    @result.reaction = @user.add_reaction(@resource)
    @result
  end

  private

  def result_class
    AddReactionServiceResult
  end
end
