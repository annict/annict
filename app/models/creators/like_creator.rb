# typed: false
# frozen_string_literal: true

module Creators
  class LikeCreator
    attr_accessor :like

    def initialize(user:, likeable:)
      @user = user
      @likeable = likeable
    end

    def call
      like = @user.likes.find_by_resource!(@likeable)

      if like
        self.like = like
        return self
      end

      ActiveRecord::Base.transaction do
        self.like = @user.like!(@likeable)
        send_notification
      end

      self
    end

    private

    def send_notification
      return if @user.id == @likeable.user_id

      like.send_notification_to(@user)
    end
  end
end
