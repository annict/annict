# frozen_string_literal: true

class AddReactionService < ApplicationService
  class ServiceResult < BaseServiceResult
    attr_accessor :reaction
  end

  def initialize(user:, resource:)
    super
    @user = user
    @resource = resource
  end

  def call
    @result.reaction = @user.add_reaction(@resource)
    @result
  end
end
