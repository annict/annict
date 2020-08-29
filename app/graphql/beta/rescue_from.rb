# frozen_string_literal: true

module Beta
  class RescueFrom
    def initialize(resolve_func)
      @resolve_func = resolve_func
    end

    def call(obj, args, ctx)
      ActiveRecord::Base.transaction do
        @resolve_func.call(obj, args, ctx)
      end
    rescue => err
      message = case err
      when ActiveRecord::RecordNotFound
        "Couldn't find #{err.model} with #{err.id}"
      else
        err.message
      end

      GraphQL::ExecutionError.new(message)
    end
  end
end
