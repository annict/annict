# frozen_string_literal: true

module Annict
  class Result < Struct
    module Types
      include Dry.Types(default: :strict)
    end

    class Error < Dry::Struct
      schema schema.strict
      attribute :message, Types::String
    end

    def self.new(*members)
      super(*members, :errors, keyword_init: true)
    end

    def self.success(*values)
      new(*values)
    end

    def self.failure(messages)
      new(errors: messages.map { |message| Error.new(message: message) })
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
