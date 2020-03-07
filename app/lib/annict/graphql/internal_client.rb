# frozen_string_literal: true

module Annict
  module GraphQL
    class InternalClient
      def initialize(viewer:)
        @viewer = viewer
      end

      def execute(query, variables: {})
        Canary::AnnictSchema.execute(query, variables: variables, context: context)
      end

      private

      attr_reader :viewer

      def context
        {
          writable: true,
          admin: true,
          viewer: viewer
        }
      end
    end
  end
end
