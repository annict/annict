# frozen_string_literal: true

module Resolvers
  class Programs
    def call(obj, args, _ctx)
      @obj = obj
      @args = args
      @collection = obj.programs
      from_arguments
    end

    private

    def from_arguments
      %i(
        unwatched
      ).each do |arg_name|
        next if @args[arg_name].blank?
        @collection = send(arg_name.to_s.underscore)
      end

      @collection = @collection.work_published.episode_published

      if @args[:orderBy].present?
        direction = @args[:orderBy][:direction]

        @collection = case @args[:orderBy][:field]
        when "STARTED_AT"
          @collection.order(started_at: direction)
        end
      end

      @collection
    end

    def unwatched
      @args[:unwatched].present? ? @collection.unwatched_all : @collection.all
    end
  end
end
