# frozen_string_literal: true

module Resolvers
  class Episodes
    def call(obj, args, _ctx)
      @obj = obj
      @args = args
      @collection = obj.episodes.all.published
      from_arguments
    end

    private

    def from_arguments
      if @args[:orderBy].present?
        direction = @args[:orderBy][:direction]

        @collection = case @args[:orderBy][:field]
        when "CREATED_AT"
          @collection.order(created_at: direction)
        when "SORT_NUMBER"
          @collection.order(sort_number: direction)
        end
      end

      @collection
    end
  end
end
