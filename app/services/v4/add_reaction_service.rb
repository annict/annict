# frozen_string_literal: true

module V4
  class AddReactionService < V4::ApplicationService
    class AddReactionServiceResult < Result
      attr_accessor :reaction
    end

    def initialize(user:, reactable:, content: :heart)
      super()
      @user = user
      @reactable = reactable
      @content = content
    end

    def call
      like = @user.likes.find_by_resource(@reactable)

      if like
        @result.reaction = like
        return @result
      end

      ActiveRecord::Base.transaction do
        @result.reaction = @user.add_reaction(@reactable, content: @content)
        send_notification
      end

      @result
    end

    private

    def send_notification
      return if @user.id == @reactable.user_id

      @result.reaction.send_notification_to(@user)
    end

    def result_class
      AddReactionServiceResult
    end
  end
end
