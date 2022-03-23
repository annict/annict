# frozen_string_literal: true

module Annict
  module Errors
    class AnnictError < StandardError; end
    class InvalidAPITokenScopeError < AnnictError; end
    class ModelMismatchError < AnnictError; end
    class NotLikeableError < AnnictError; end
    class UnknownImageFieldError < AnnictError; end
  end
end
