# frozen_string_literal: true

module Types
  module Objects
    class Base < GraphQL::Schema::Object
      def annict_id
        object.id
      end
    end
  end
end
