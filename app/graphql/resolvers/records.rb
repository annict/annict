# frozen_string_literal: true

module Resolvers
  class Records
    def call(obj, args, _ctx)
      @obj = obj
      @args = args
      @collection = obj.records.all
      from_arguments
    end

    private

    def from_arguments
      if @args[:orderBy].present?
        direction = @args[:orderBy][:direction]

        @collection = case @args[:orderBy][:field]
        when "CREATED_AT"
          @collection.order(created_at: direction)
        when "LIKES_COUNT"
          @collection.order(likes_count: direction)
        end
      end

      @collection
    end
  end
end
