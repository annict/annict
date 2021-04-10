# frozen_string_literal: true

module Annict
  class Result < Struct
    def self.new(*members)
      super(*members, :errors, keyword_init: true)
    end

    def self.success(*values)
      new(*values)
    end

    # @param errors [ActiveModel::Errors]
    def self.failure(errors)
      new(errors: errors)
    end

    def success?
      errors.blank?
    end

    def failure?
      !success?
    end

    private

    def initialize(*args)
      values = { errors: [] }.merge(args.first || {})
      super(**values)
    end
  end
end
