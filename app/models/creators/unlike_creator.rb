# frozen_string_literal: true

module Creators
  class UnlikeCreator
    def initialize(user:, likeable:)
      @user = user
      @likeable = likeable
    end

    def call
      @user.unlike(@likeable)

      self
    end
  end
end
