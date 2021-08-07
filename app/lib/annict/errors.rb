# frozen_string_literal: true

module Annict
  module Errors
    class AnnictError < StandardError; end

    class InvalidAPITokenScopeError < AnnictError; end

    class ModelMismatchError < AnnictError; end

    class NotLikeableError < AnnictError; end
  end
end
