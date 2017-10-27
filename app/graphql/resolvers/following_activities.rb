# frozen_string_literal: true

module Resolvers
  class FollowingActivities
    def call(obj, args, _ctx)
      @obj = obj
      @args = args
      following_ids = obj.followings.pluck(:id)
      following_ids << obj.id
      @collection = Activity.where(user_id: following_ids)
      from_arguments
    end

    private

    def from_arguments
      if @args[:orderBy].present?
        direction = @args[:orderBy][:direction]

        @collection = case @args[:orderBy][:field]
        when "CREATED_AT"
          @collection.order(created_at: direction)
        end
      end

      @collection
    end
  end
end
